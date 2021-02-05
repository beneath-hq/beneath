# allows us to use Client as a type hint without an import cycle
# see: https://www.stefaanlippens.net/circular-imports-type-hints-python.html
# pylint: disable=wrong-import-position,ungrouped-imports
from __future__ import annotations
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    pass

import asyncio
from datetime import timedelta
from enum import Enum
import json
import os
import signal
import sys
from typing import Any, Dict, List, Set, Tuple
import uuid

import msgpack

from beneath import config
from beneath.client import Client
from beneath.instance import StreamInstance
from beneath.logging import logger
from beneath.pipeline.parse_args import parse_pipeline_args
from beneath.stream import Stream
from beneath.utils import AIODelayBuffer, ServiceQualifier, StreamQualifier
from beneath.writer import Writer


class Action(Enum):
    test = "test"
    stage = "stage"
    run = "run"
    teardown = "teardown"


class Strategy(Enum):
    continuous = "continuous"
    delta = "delta"
    batch = "batch"


SERVICE_CHECKPOINT_LOG_RETENTION = timedelta(hours=6)
SERVICE_CHECKPOINT_SCHEMA = """
  type ServiceCheckpoint @stream @key(fields: ["key"]) {
    key: String!
    value: Bytes
  }
"""


class BasePipeline:

    instances: Dict[StreamQualifier, StreamInstance]
    checkpoint_instance: StreamInstance
    writer: Writer
    checkpoint_writer: BasePipeline._CheckpointWriter
    service_id: uuid.UUID
    description: str

    _stage_tasks: List[asyncio.Task]
    _input_qualifiers: Set[StreamQualifier]
    _output_qualifiers: Set[StreamQualifier]

    # INITIALIZATION AND OPTIONS PARSING

    def __init__(
        self,
        action: Action = None,
        strategy: Strategy = None,
        version: int = None,
        service_path: str = None,
        service_read_quota: int = None,
        service_write_quota: int = None,
        service_scan_quota: int = None,
        parse_args: bool = False,
        client: Client = None,
        write_delay_ms: int = config.DEFAULT_WRITE_DELAY_MS,
        write_checkpoint_delay_ms: int = config.DEFAULT_WRITE_PIPELINE_CHECKPOINT_DELAY_MS,
    ):
        if parse_args:
            parse_pipeline_args()
        self.action = self._arg_or_env("action", action, "BENEATH_PIPELINE_ACTION", fn=Action)
        self.strategy = self._arg_or_env(
            "strategy", strategy, "BENEATH_PIPELINE_STRATEGY", fn=Strategy
        )
        self.version = self._arg_or_env("version", version, "BENEATH_PIPELINE_VERSION", fn=int)
        self.service_qualifier = self._arg_or_env(
            "service_path",
            service_path,
            "BENEATH_PIPELINE_SERVICE_PATH",
            fn=ServiceQualifier.from_path,
        )
        self.service_read_quota = self._arg_or_env(
            "service_read_quota",
            service_read_quota,
            "BENEATH_PIPELINE_SERVICE_READ_QUOTA",
            fn=int,
            default=lambda: None,
        )
        self.service_write_quota = self._arg_or_env(
            "service_write_quota",
            service_write_quota,
            "BENEATH_PIPELINE_SERVICE_WRITE_QUOTA",
            fn=int,
            default=lambda: None,
        )
        self.service_scan_quota = self._arg_or_env(
            "service_scan_quota",
            service_scan_quota,
            "BENEATH_PIPELINE_SERVICE_SCAN_QUOTA",
            fn=int,
            default=lambda: None,
        )
        self.client = client if client else Client()
        self.write_delay_ms = write_delay_ms
        self.write_checkpoint_delay_ms = write_checkpoint_delay_ms
        self.instances: Dict[StreamQualifier, StreamInstance] = {}
        self.checkpoint_instance = None
        self.writer = None
        self.checkpoint_writer = None
        self.service_id = None
        self.description = None
        self.source_url = None
        self._stage_tasks = []
        self._input_qualifiers = set()
        self._output_qualifiers = set()
        self.logger = logger
        self.dry = self.action == Action.test
        self._run_task: asyncio.Task = None
        self._init()

    @staticmethod
    def _arg_or_env(name, arg, env, fn=None, default=None):
        if arg is not None:
            if fn:
                return fn(arg)
            return arg
        arg = os.getenv(env)
        if arg is not None:
            if fn:
                return fn(arg)
            return arg
        if default:
            if callable(default):
                return default()
            return default
        raise ValueError(f"Must provide {name} arg or set {env} environment variable")

    # OVERRIDES

    def _init(self):  # pylint: disable=no-self-use
        raise Exception("_init should be implemented in a subclass")

    async def _run(self):
        raise Exception("_run should be implemented in a subclass")

    async def _before_run(self):
        raise Exception("_before_run should be implemented in a subclass")

    async def _after_run(self):
        raise Exception("_after_run should be implemented in a subclass")

    # RUNNING

    def main(self):
        loop = asyncio.get_event_loop()

        async def _main():
            try:
                self._run_task = asyncio.create_task(self.run())
                await self._run_task
            except asyncio.CancelledError:
                logger.info("Pipeline gracefully cancelled")

        async def _kill(sig):
            logger.info("Received %s, attempting graceful shutdown...", sig.name)
            if self._run_task:
                self._run_task.cancel()

        signals = (signal.SIGTERM, signal.SIGINT)
        for sig in signals:
            loop.add_signal_handler(sig, lambda sig=sig: asyncio.create_task(_kill(sig)))

        loop.run_until_complete(_main())

    async def run(self):
        await self._stage()
        if self.action == Action.stage:
            return

        if self.action == Action.teardown:
            await self._teardown()
            return

        if self.dry:
            logger.info("Running in TEST mode: Records will be printed, but NOT saved to Beneath")
        else:
            logger.info("Running in PRODUCTION mode: Records will be saved to Beneath")

        self.writer = self.client.writer(dry=self.dry, write_delay_ms=self.write_delay_ms)
        self.checkpoint_writer = self._CheckpointWriter(self)

        await self._before_run()
        await self.writer.start()
        await self.checkpoint_writer.start()
        try:
            await self._run()
            await self.checkpoint_writer.stop()
            await self.writer.stop()
            await self._after_run()
        except asyncio.CancelledError:
            logger.info("Pipeline cancelled, flushing buffered records...")
            await self.checkpoint_writer.stop()
            await self.writer.stop()
            raise

        logger.info("Pipeline completed successfully")

    # SERVICE CHECKPOINT

    class _CheckpointWriter(AIODelayBuffer[Tuple[str, Any]]):

        checkpoint: Dict[str, Any]

        def __init__(self, pipeline: BasePipeline):
            super().__init__(
                max_delay_ms=pipeline.write_checkpoint_delay_ms,
                max_size=sys.maxsize,
                max_count=sys.maxsize,
            )
            self.pipeline = pipeline

        def _reset(self):
            self.checkpoint = {}

        def _merge(self, value: Tuple[str, Any]):
            (key, checkpoint) = value
            self.checkpoint[key] = checkpoint

        async def _flush(self):
            records = (
                {
                    "key": key,
                    "value": msgpack.packb(checkpoint, datetime=True),
                }
                for (key, checkpoint) in self.checkpoint.items()
            )
            await self.pipeline.writer.write(
                instance=self.pipeline.checkpoint_instance, records=records
            )

        # pylint: disable=arguments-differ
        async def write(self, key: str, checkpoint: Any):
            await super().write(value=(key, checkpoint), size=0)

    async def get_checkpoint(self, key: str, default: Any = None) -> Any:
        if self.dry:
            return default

        cursor = await self.checkpoint_instance.query_index(
            filter=json.dumps(
                {
                    "key": key,
                }
            )
        )

        value = default
        batch = await cursor.read_next(limit=1)
        if batch is not None:
            records = list(batch)
            if len(records) > 0:
                value = records[0]["value"]
                value = msgpack.unpackb(value, timestamp=3)

        return value

    async def set_checkpoint(self, key: str, value: Any):
        await self.checkpoint_writer.write(key, value)

    # STAGING

    @property
    def is_stage(self):
        return (self.action == Action.stage) or (self.action == Action.teardown)

    async def _stage(self):
        logger.info("Staging service '%s'", self.service_qualifier)
        await self._stage_service()
        await asyncio.gather(self._stage_service_checkpoint(), *self._stage_tasks)
        logger.info("Finished staging pipeline for service '%s'", self.service_qualifier)

    async def _stage_service(self):
        if self.is_stage:
            description = self.description if self.description else "Service for running pipeline"
            admin_data = await self.client.admin.services.create(
                organization_name=self.service_qualifier.organization,
                project_name=self.service_qualifier.project,
                service_name=self.service_qualifier.service,
                description=description,
                source_url=self.source_url,
                read_quota_bytes=self.service_read_quota,
                write_quota_bytes=self.service_write_quota,
                scan_quota_bytes=self.service_scan_quota,
                update_if_exists=True,
            )
        else:
            admin_data = await self.client.admin.services.find_by_organization_project_and_name(
                organization_name=self.service_qualifier.organization,
                project_name=self.service_qualifier.project,
                service_name=self.service_qualifier.service,
            )
        self.service_id = uuid.UUID(hex=admin_data["serviceID"])

    async def _stage_service_checkpoint(self):
        stream_name = f"{self.service_qualifier.service}-checkpoint"
        qualifier = StreamQualifier(
            organization=self.service_qualifier.organization,
            project=self.service_qualifier.project,
            stream=stream_name,
        )
        if self.is_stage:
            description = (
                "Stores pipeline checkpoints for service "
                f"'{self.service_qualifier.service}' (automatically created)"
            )
            stream = await self.client.create_stream(
                stream_path=str(qualifier),
                schema=SERVICE_CHECKPOINT_SCHEMA,
                description=description,
                log_retention=SERVICE_CHECKPOINT_LOG_RETENTION,
                use_warehouse=False,
                update_if_exists=True,
            )
            await self._update_service_permissions(stream, read=True, write=True)
        else:
            stream = await self.client.find_stream(stream_path=str(qualifier))
        instance = await stream.create_instance(
            version=self.version,
            dry=self.dry,
            update_if_exists=True,
        )
        self.checkpoint_instance = instance
        logger.info(
            "Staged stream for pipeline checkpoint '%s' (using version %i)", qualifier, self.version
        )

    def _parse_stream_path(self, stream_path: str) -> StreamQualifier:
        # use service organization and project if "stream_path" is just a name
        if "/" not in stream_path:
            stream_path = (
                f"{self.service_qualifier.organization}/"
                + f"{self.service_qualifier.project}/"
                + stream_path
            )
        return StreamQualifier.from_path(stream_path)

    def _add_input_stream(self, stream_path: str) -> StreamQualifier:
        qualifier = self._parse_stream_path(stream_path)
        if (qualifier in self._input_qualifiers) or (qualifier in self._output_qualifiers):
            raise ValueError(f"Stream {qualifier} already added as input or output")
        self._input_qualifiers.add(qualifier)
        self._stage_tasks.append(self._stage_input_stream(str(qualifier)))
        return qualifier

    async def _stage_input_stream(self, stream_path: str):
        stream = await self.client.find_stream(stream_path=stream_path)
        instance = stream.primary_instance
        if not instance:
            raise ValueError(f"Input stream {stream_path} doesn't have a primary instance")
        assert stream.qualifier not in self.instances
        if self.is_stage:
            await self._update_service_permissions(stream, read=True)
        self.instances[stream.qualifier] = stream.primary_instance
        logger.info(
            "Staged input stream '%s' (using version %i)",
            stream_path,
            stream.primary_instance.version,
        )

    def _add_output_stream(
        self,
        stream_path: str,
        schema: str = None,
        description: str = None,
        use_index: bool = None,
        use_warehouse: bool = None,
        retention: timedelta = None,
        log_retention: timedelta = None,
        index_retention: timedelta = None,
        warehouse_retention: timedelta = None,
    ) -> StreamQualifier:
        qualifier = self._parse_stream_path(stream_path)
        if qualifier in self._output_qualifiers:
            return qualifier
        if qualifier in self._input_qualifiers:
            raise ValueError(f"Cannot use stream '{qualifier}' as both input and output")
        self._output_qualifiers.add(qualifier)
        self._stage_tasks.append(
            self._stage_output_stream(
                stream_path=str(qualifier),
                schema=schema,
                description=description,
                use_index=use_index,
                use_warehouse=use_warehouse,
                retention=retention,
                log_retention=log_retention,
                index_retention=index_retention,
                warehouse_retention=warehouse_retention,
            )
        )
        return qualifier

    async def _stage_output_stream(
        self,
        stream_path: str,
        schema: str = None,
        description: str = None,
        use_index: bool = None,
        use_warehouse: bool = None,
        retention: timedelta = None,
        log_retention: timedelta = None,
        index_retention: timedelta = None,
        warehouse_retention: timedelta = None,
    ):
        if not schema:
            if description:
                raise ValueError("you cannot set a stream 'description' without setting 'schema'")
            stream = await self.client.find_stream(stream_path)
        else:
            if self.is_stage:
                stream = await self.client.create_stream(
                    stream_path=stream_path,
                    schema=schema,
                    description=description,
                    use_index=use_index,
                    use_warehouse=use_warehouse,
                    retention=retention,
                    log_retention=log_retention,
                    index_retention=index_retention,
                    warehouse_retention=warehouse_retention,
                    update_if_exists=True,
                )
            else:
                stream = await self.client.find_stream(stream_path=stream_path)
        assert stream.qualifier not in self.instances
        instance = await stream.create_instance(
            version=self.version, dry=self.dry, update_if_exists=True
        )
        if self.is_stage:
            await self._update_service_permissions(stream, write=True)
        self.instances[stream.qualifier] = instance
        logger.info("Staged output stream '%s' (using version %s)", stream_path, self.version)

    async def _update_service_permissions(
        self, stream: Stream, read: bool = None, write: bool = None
    ):
        await self.client.admin.services.update_permissions_for_stream(
            service_id=self.service_id,
            stream_id=stream.stream_id,
            read=read,
            write=write,
        )

    # TEARDOWN

    async def _teardown(self):
        logger.info("Tearing down pipeline for service '%s'", self.service_qualifier)
        for qualifier in self._output_qualifiers:
            instance = self.instances[qualifier]
            await instance.stream.delete()
            logger.info("Deleted output stream '%s'", qualifier)
        await self.checkpoint_instance.stream.delete()
        logger.info(
            "Deleted stream for pipeline checkpoint '%s'", self.checkpoint_instance.stream.qualifier
        )
        await self.client.admin.services.delete(self.service_id)
        logger.info("Deleted pipeline service '%s'", self.service_qualifier)
        logger.info("Finished teardown of pipeline for service '%s'", self.service_qualifier)
