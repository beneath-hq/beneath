# allows us to use Client as a type hint without an import cycle
# see: https://www.stefaanlippens.net/circular-imports-type-hints-python.html
# pylint: disable=wrong-import-position,ungrouped-imports
from __future__ import annotations
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    pass

import asyncio
from collections import defaultdict
from collections.abc import Mapping
from datetime import timedelta
from typing import AsyncIterator, Awaitable, Callable, Dict, Iterable, List, Union

from beneath.config import DEFAULT_READ_BATCH_SIZE
from beneath.cursor import Cursor
from beneath.instance import StreamInstance
from beneath.logging import logger
from beneath.pipeline import BasePipeline, Strategy
from beneath.utils import StreamQualifier

AsyncGenerateFn = Callable[[BasePipeline], AsyncIterator[Mapping]]
AsyncApplyFn = Callable[[Mapping], Awaitable[Union[Mapping, Iterable[Mapping]]]]

PIPELINE_IDLE = "<PIPELINE_IDLE>"


class Transform:
    def __init__(self, pipeline: Pipeline):
        self.pipeline = pipeline

    # pylint: disable=no-self-use
    async def process(self, incoming_records: Iterable[Mapping]):
        raise Exception("Must implement process in subclass")


class Generate(Transform):
    def __init__(self, pipeline: Pipeline, fn: AsyncGenerateFn):
        super().__init__(pipeline)
        self.fn = fn

    async def process(self, incoming_records: Iterable[Mapping]):
        assert incoming_records is None
        async for records in self.fn(self.pipeline):
            if records == PIPELINE_IDLE:
                if self.pipeline.strategy != Strategy.continuous:
                    return
                continue
            if not isinstance(records, list):
                records = [records]
            yield records


class ReadStream(Transform):

    qualifier: StreamQualifier
    limit: int
    instance: StreamInstance
    checkpoint_key: str
    cursor: Cursor

    def __init__(
        self, pipeline: Pipeline, stream_path: str, batch_size: int = DEFAULT_READ_BATCH_SIZE
    ):
        super().__init__(pipeline)
        self.qualifier = self.pipeline._add_input_stream(stream_path)
        self.limit = batch_size

    async def process(self, incoming_records: Iterable[Mapping]):
        assert incoming_records is None
        await self._init_cursor()
        iterators = [self._replay()]
        if self.pipeline.strategy == Strategy.delta:
            iterators.append(self._changes_delta())
        elif self.pipeline.strategy == Strategy.continuous:
            iterators.append(self._changes_continuous())
        for it in iterators:
            async for batch in it:
                yield batch
                await self._write_checkpoint()

    async def _init_cursor(self):
        self.instance = self.pipeline.instances[self.qualifier]
        self.checkpoint_key = str(self.instance.instance_id) + ":cursor"
        checkpoint = await self.pipeline.get_checkpoint(self.checkpoint_key)
        if checkpoint:
            self.cursor = Cursor(
                connection=self.instance.stream.client.connection,
                schema=self.instance.stream.schema,
                replay_cursor=checkpoint.get("replay"),
                changes_cursor=checkpoint.get("changes"),
            )
        else:
            self.cursor = await self.instance.query_log()

    async def _write_checkpoint(self):
        checkpoint = {}
        if self.cursor.replay_cursor:
            checkpoint["replay"] = self.cursor.replay_cursor
        if self.cursor.changes_cursor:
            checkpoint["changes"] = self.cursor.changes_cursor
        await self.pipeline.set_checkpoint(self.checkpoint_key, checkpoint)

    async def _replay(self):
        if not self.cursor.replay_cursor:
            return
        while True:
            batch = await self.cursor.read_next(limit=self.limit)
            if not batch:
                return
            yield batch

    async def _changes_delta(self):
        if not self.cursor.changes_cursor:
            return
        while True:
            batch = await self.cursor.read_next_changes(limit=self.limit)
            if not batch:
                return
            batch = list(batch)
            if len(batch) == 0:
                return
            yield batch

    async def _changes_continuous(self):
        if not self.cursor.changes_cursor:
            return
        async for batch in self.cursor.subscribe_changes(batch_size=self.limit):
            yield batch


class Apply(Transform):
    def __init__(self, pipeline: Pipeline, fn: AsyncApplyFn, max_concurrency: int = None):
        super().__init__(pipeline)
        self.fn = fn
        self.max_concurrency = (
            max_concurrency if max_concurrency else 1
        )  # TODO: use max_concurrency :(

    async def process(self, incoming_records: Iterable[Mapping]):
        for record in incoming_records:
            fut = self.fn(record)
            if hasattr(fut, "__aiter__"):
                async for record in fut:
                    yield [record]
            else:
                batch = await fut
                if batch is None:
                    continue
                if isinstance(batch, list):
                    yield batch
                else:
                    yield [batch]


class WriteStream(Transform):
    def __init__(
        self,
        pipeline: Pipeline,
        stream_path: str,
        schema: str = None,
        description: str = None,
        retention: timedelta = None,
    ):
        super().__init__(pipeline)
        self.qualifier = self.pipeline._add_output_stream(
            stream_path, schema, description=description, retention=retention
        )

    async def process(self, incoming_records: Iterable[Mapping]):
        instance = self.pipeline.instances[self.qualifier]
        await self.pipeline.writer.write(
            instance=instance,
            records=incoming_records,
        )
        # trick to turn into async iterator
        return
        # pylint: disable=unreachable
        yield


class Pipeline(BasePipeline):

    _initial: List[Transform]
    _dag: Dict[Transform, List[Transform]]

    # INIT

    def _init(self):
        self._initial = []
        self._dag = defaultdict(list)

    # DAG

    def generate(self, fn: AsyncGenerateFn) -> Transform:
        transform = Generate(self, fn)
        self._initial.append(transform)
        return transform

    def read_stream(self, stream_path: str) -> Transform:
        transform = ReadStream(self, stream_path)
        self._initial.append(transform)
        return transform

    def apply(
        self, prev_transform: Transform, fn: AsyncApplyFn, max_concurrency: int = None
    ) -> Transform:
        transform = Apply(self, fn, max_concurrency)
        self._dag[prev_transform].append(transform)
        return transform

    def write_stream(
        self,
        prev_transform: Transform,
        stream_path: str,
        schema: str = None,
        description: str = None,
        retention: timedelta = None,
    ):
        transform = WriteStream(
            self, stream_path, schema, description=description, retention=retention
        )
        self._dag[prev_transform].append(transform)

    # RUNNING

    async def _run(self):
        await asyncio.gather(*[self._run_transform(transform, None) for transform in self._initial])

    async def _run_transform(self, transform: Transform, incoming_records: List):
        async for outgoing_records in transform.process(incoming_records):
            tasks = (
                self._run_transform(next_transform, outgoing_records)
                for next_transform in self._dag[transform]
            )
            await asyncio.gather(*tasks)

    async def _before_run(self):
        if self.strategy == Strategy.batch:
            return
        if not self.checkpoint_instance.is_primary:
            await self.checkpoint_instance.update(make_primary=True)
        for qualifier in self._output_qualifiers:
            instance = self.instances[qualifier]
            if not instance.is_primary:
                await instance.update(make_primary=True)
                logger.info(
                    "Made version %s of output stream '%s' primary", instance.version, qualifier
                )
            else:
                logger.info(
                    "Using existing primary version %s for output stream '%s'",
                    instance.version,
                    qualifier,
                )

    async def _after_run(self):
        if self.strategy != Strategy.batch:
            return
        if not self.checkpoint_instance.is_primary or not self.checkpoint_instance.is_final:
            await self.checkpoint_instance.update(make_primary=True, make_final=True)
        for qualifier in self._output_qualifiers:
            instance = self.instances[qualifier]
            if not instance.is_primary or not instance.is_final:
                await instance.update(make_primary=True, make_final=True)
                logger.info(
                    "Made version %s for output stream '%s' final and primary",
                    instance.version,
                    qualifier,
                )
