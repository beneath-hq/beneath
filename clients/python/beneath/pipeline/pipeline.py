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
from beneath.pipeline import BasePipeline, Strategy

AsyncGenerateFn = Callable[[BasePipeline], AsyncIterator[Mapping]]
AsyncApplyFn = Callable[[Mapping], Awaitable[Union[Mapping, Iterable[Mapping]]]]

PIPELINE_IDLE = "<PIPELINE_IDLE>"


class Transform:
    def __init__(self, pipeline: Pipeline):
        self.pipeline = pipeline

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
    def __init__(
        self,
        pipeline: Pipeline,
        stream_path: str,
        batch_size: int = DEFAULT_READ_BATCH_SIZE,
    ):
        super().__init__(pipeline)
        self.stream_path = stream_path
        self.batch_size = batch_size
        self.qualifier = self.pipeline._add_input_stream(stream_path)

    async def process(self, incoming_records: Iterable[Mapping]):
        assert incoming_records is None
        consumer = self.pipeline._inputs[self.qualifier]
        async for batch in consumer.iterate(
            batches=True,
            replay_only=(self.pipeline.strategy == Strategy.batch),
            stop_when_idle=(self.pipeline.strategy == Strategy.delta),
        ):
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
            stream_path,
            schema=schema,
            description=description,
            retention=retention,
        )

    async def process(self, incoming_records: Iterable[Mapping]):
        instance = self.pipeline._outputs[self.qualifier]
        await self.pipeline.client.write(
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
        self,
        prev_transform: Transform,
        fn: AsyncApplyFn,
        max_concurrency: int = None,
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
            self,
            stream_path,
            schema,
            description=description,
            retention=retention,
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
        # make outputs primary if not batch
        if self.strategy == Strategy.batch:
            return
        for qualifier, instance in self._outputs.items():
            if not instance.is_primary:
                await instance.update(make_primary=True)
                self.logger.info(
                    "Made version %s of output stream '%s' primary",
                    instance.version,
                    qualifier,
                )
            else:
                self.logger.info(
                    "Using existing primary version %s for output stream '%s'",
                    instance.version,
                    qualifier,
                )

    async def _after_run(self):
        # make outputs primary if batch
        if self.strategy != Strategy.batch:
            return
        for qualifier, instance in self._outputs.items():
            if not instance.is_primary or not instance.is_final:
                await instance.update(make_primary=True, make_final=True)
                self.logger.info(
                    "Made version %s for output stream '%s' final and primary",
                    instance.version,
                    qualifier,
                )
