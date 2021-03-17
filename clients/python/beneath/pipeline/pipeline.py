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

from beneath import config
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
        batch_size: int = config.DEFAULT_READ_BATCH_SIZE,
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
        schema_kind: str = "GraphQL",
        description: str = None,
        retention: timedelta = None,
    ):
        super().__init__(pipeline)
        self.qualifier = self.pipeline._add_output_stream(
            stream_path,
            schema=schema,
            schema_kind=schema_kind,
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
    """
    Pipelines are a construct built on top of the Beneath primitives to manage the reading of input
    streams, creation of output streams, data generation and derivation logic, and more.

    This simple implementation supports four combinatorial operations: ``generate``,
    ``read_stream``, ``apply``, and ``write_stream``. It's suitable for generating streams,
    consuming streams, and one-to-N derivation of one stream to another. It's not currently
    suitable for advanced aggregation or multi-machine parallel processing.

    It supports (light) stateful transformations via a key-value based `checkpointer`, which is
    useful for tracking generator and consumer progress in between invocations (or crashes).

    Args:
        action (Action):
            The action to run when calling ``main`` or ``execute``
        strategy (Strategy):
            The processing strategy to apply when action="test" or action="run"
        version (int):
            The version number for output streams. Incrementing the version number will cause
            the pipeline to create new output stream instances and replay input streams.
        service_path (str):
            Path for a service to create for the pipeline when action="stage". The service will
            be assigned correct permissions for reading input streams and writing to output streams
            and checkpoints. The service can be used to create a secret for deploying the pipeline
            to production.
        service_read_quota (int):
            Read quota for the service staged for the pipeline (see ``service_path``)
        service_write_quota (int):
            Write quota for the service staged for the pipeline (see ``service_path``)
        service_scan_quota (int):
            Scan quota for the service staged for the pipeline (see ``service_path``)
        client (Client):
            Client to use for the pipeline. If not set, initializes a new client (see its init
            for details on which secret gets used).
        write_delay_ms (int):
            Passed to ``Client`` initializer if ``client`` arg isn't passed
        disable_checkpoints (bool):
            If true, will not create a checkpointer, and consumers will not save state.
            Defaults to false.

    Example::

        # pipeline.py

        # This pipeline generates a stream `ticks` with a record for every minute
        # since 1st Jan 2021.

        # To test locally:
        #     python ./pipeline.py test
        # To prepare for running:
        #     python ./pipeline.py stage USERNAME/PROJECT/ticker
        # To run (after stage):
        #     python ./pipeline.py run USERNAME/PROJECT/ticker
        # To teardown:
        #     python ./pipeline.py teardown USERNAME/PROJECT/ticker

        import asyncio
        import beneath
        from datetime import datetime, timedelta, timezone

        start = datetime(year=2021, month=1, day=1, tzinfo=timezone.utc)

        async def ticker(p: beneath.Pipeline):
            last_tick = await p.checkpoints.get("last", default=start)
            while True:
                now = datetime.now(tz=last_tick.tzinfo)
                next_tick = last_tick + timedelta(minutes=1)

                if next_tick >= now:
                    yield beneath.PIPELINE_IDLE
                    wait = next_tick - now
                    await asyncio.sleep(wait.total_seconds())

                yield {"time": next_tick}
                await p.checkpoints.set("last", next_tick)
                last_tick = next_tick

        if __name__ == "__main__":

            p = beneath.Pipeline(parse_args=True)
            p.description = "Pipeline that emits a tick for every minute since 1st Jan 2021"

            ticks = p.generate(ticker)
            p.write_stream(
                ticks,
                stream_path="ticks",
                schema='''
                    type Tick @schema {
                        time: Timestamp! @key
                    }
                ''',
            )

            p.main()
    """

    _initial: List[Transform]
    _dag: Dict[Transform, List[Transform]]

    # INIT OVERRIDE

    def _init(self):
        self._initial = []
        self._dag = defaultdict(list)

    # DAG

    def generate(self, fn: AsyncGenerateFn) -> Transform:
        """
        Pipeline step for generating records.

        Args:
            fn (async def fn(pipeline)):
                An async iterator that generates records. The pipeline is passed as an input arg.
        """
        transform = Generate(self, fn)
        self._initial.append(transform)
        return transform

    def read_stream(self, stream_path: str) -> Transform:
        """
        Pipeline step for consuming the primary instance of a stream.

        Args:
            stream_path (str):
                The stream to consume
        """
        transform = ReadStream(self, stream_path)
        self._initial.append(transform)
        return transform

    def apply(
        self,
        prev_transform: Transform,
        fn: AsyncApplyFn,
        max_concurrency: int = None,
    ) -> Transform:
        """
        Pipeline step that transforms records emitted by a previous step.

        Args:
            prev_transform (Transform):
                The pipeline step to apply on. Can be a `generate`, `read_stream`, or other
                `apply` step.
            fn (async def fn(record) -> (None | record | [record])):
                Function applied to each incoming record. Can return ``None``, one record
                or a list of records, which will propagate to the next pipeline step.
        """
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
        schema_kind: str = "GraphQL",
    ):
        """
        Pipeline step that writes incoming records from the previous step to a stream.

        Args:
            prev_transform (Transform):
                The pipeline step to apply on. Can be a `generate`, `read_stream`, or other
                `apply` step.
            stream_path (str):
                The stream to output to. If ``schema`` is provided, the stream will be created when
                running the ``stage`` action. If the path doesn't include a username and project
                name, it will attempt to find or create the stream in the pipeline's service's
                project (see ``service_path`` in the constructor).
            schema (str):
                A GraphQL schema for creating the output stream.
            description (str):
                A description for the stream (only applicable if schema is set).
            retention (timedelta):
                The amount of time to retain written data in the stream. By default, records
                are saved forever.
            schema_kind (str):
                The parser to use for ``schema``. Currently must be "GraphQL" (default).
        """
        transform = WriteStream(
            self,
            stream_path,
            schema,
            description=description,
            retention=retention,
            schema_kind=schema_kind,
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
