# allows us to use Client as a type hint without an import cycle
# see: https://www.stefaanlippens.net/circular-imports-type-hints-python.html
# pylint: disable=wrong-import-position,ungrouped-imports
from __future__ import annotations
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    pass

import asyncio
from enum import Enum
import os
import signal
from typing import Dict, Set
import uuid

from beneath import config
from beneath.connection import GraphQLError
from beneath.consumer import Consumer
from beneath.client import Client
from beneath.checkpointer import Checkpointer
from beneath.instance import StreamInstance
from beneath.pipeline.parse_args import parse_pipeline_args
from beneath.stream import Stream
from beneath.utils import ServiceQualifier, StreamQualifier


class Action(Enum):
    test = "test"
    stage = "stage"
    run = "run"
    teardown = "teardown"


class Strategy(Enum):
    continuous = "continuous"
    delta = "delta"
    batch = "batch"


class BasePipeline:
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
    ):
        if parse_args:
            parse_pipeline_args()
        self.action = self._arg_or_env("action", action, "BENEATH_PIPELINE_ACTION", fn=Action)
        self.strategy = self._arg_or_env(
            "strategy",
            strategy,
            "BENEATH_PIPELINE_STRATEGY",
            fn=Strategy,
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
        dry = self.action == Action.test
        self.client = client if client else Client(dry=dry, write_delay_ms=write_delay_ms)
        self.logger = self.client.logger
        self.checkpoints: Checkpointer = None
        self.service_id: uuid.UUID = None
        self.description: str = None

        self._inputs: Dict[StreamQualifier, Consumer] = {}
        self._input_qualifiers: Set[StreamQualifier] = set()
        self._outputs: Dict[StreamQualifier, StreamInstance] = {}
        self._output_qualifiers: Set[StreamQualifier] = set()
        self._output_kwargs: Dict[StreamQualifier, dict] = {}

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

    async def _before_run(self):
        raise Exception("_before_run should be implemented in a subclass")

    async def _after_run(self):
        raise Exception("_after_run should be implemented in a subclass")

    async def _run(self):
        raise Exception("_run should be implemented in a subclass")

    # RUNNING

    def main(self):
        """
        Runs the pipeline in an asyncio event loop and gracefully handles SIGINT and SIGTERM
        """
        loop = asyncio.get_event_loop()

        async def _main():
            try:
                self._run_task = asyncio.create_task(self.run())
                await self._run_task
            except asyncio.CancelledError:
                self.logger.info("Pipeline gracefully cancelled")

        async def _kill(sig):
            self.logger.info("Received %s, attempting graceful shutdown...", sig.name)
            if self._run_task:
                self._run_task.cancel()

        signals = (signal.SIGTERM, signal.SIGINT)
        for sig in signals:
            loop.add_signal_handler(sig, lambda sig=sig: asyncio.create_task(_kill(sig)))

        loop.run_until_complete(_main())

    async def run(self):
        """
        Runs the pipeline in accordance with the action and strategy set on init
        """
        await self.client.start()
        try:
            if self.action == Action.test:
                await self._run_test()
            elif self.action == Action.stage:
                await self._run_stage()
            elif self.action == Action.run:
                await self._run_run()
            elif self.action == Action.teardown:
                await self._run_teardown()
        except asyncio.CancelledError:
            await self.client.stop()
            self.logger.info("Pipeline cancelled")
            raise

    async def _run_test(self):
        await self._stage_checkpointer(create=True)
        await asyncio.gather(
            *[self._stage_input_stream(qualifier) for qualifier in self._input_qualifiers],
            *[
                self._stage_output_stream(qualifier=qualifier, create=True)
                for qualifier in self._output_qualifiers
            ],
        )
        self.logger.info("Running in TEST mode: Records will be printed, but not saved")
        await self._before_run()
        await self._run()
        await self._after_run()
        await self.client.stop()
        self.logger.info("Finished running pipeline")

    async def _run_stage(self):
        async def _stage_input(qualifier: StreamQualifier):
            await self._stage_input_stream(qualifier)
            consumer = self._inputs[qualifier]
            await self._update_service_permissions(consumer.instance.stream, read=True)

        async def _stage_output(qualifier: StreamQualifier):
            await self._stage_output_stream(qualifier=qualifier, create=True)
            instance = self._outputs[qualifier]
            await self._update_service_permissions(instance.stream, write=True)

        await self._stage_service(create=True)
        await self._stage_checkpointer(create=True)
        await self._update_service_permissions(
            self.checkpoints.instance.stream, read=True, write=True
        )
        await asyncio.gather(*[_stage_input(qualifier) for qualifier in self._input_qualifiers])
        await asyncio.gather(*[_stage_output(qualifier) for qualifier in self._output_qualifiers])
        await self.client.stop()
        self.logger.info("Finished staging pipeline for service '%s'", self.service_qualifier)

    async def _run_run(self):
        await self._stage_checkpointer(create=False)
        await asyncio.gather(
            *[self._stage_input_stream(qualifier) for qualifier in self._input_qualifiers],
            *[
                self._stage_output_stream(qualifier=qualifier, create=False)
                for qualifier in self._output_qualifiers
            ],
        )
        self.logger.info("Running in PRODUCTION mode: Records will be saved to Beneath")
        await self._before_run()
        await self._run()
        await self.client.stop()
        await self._after_run()
        self.logger.info("Finished running pipeline")

    async def _run_teardown(self):
        await asyncio.gather(
            *[self._teardown_output_stream(qualifier) for qualifier in self._output_qualifiers]
        )
        await self._teardown_checkpointer()
        await self._teardown_service()
        await self.client.stop()
        self.logger.info("Finished tearing down pipeline for service '%s'", self.service_qualifier)

    # STAGING

    async def _stage_service(self, create: bool):
        if create:
            description = (
                self.description
                if self.description
                else f"Service for running '{self.service_qualifier.service}' pipeline"
            )
            admin_data = await self.client.admin.services.create(
                organization_name=self.service_qualifier.organization,
                project_name=self.service_qualifier.project,
                service_name=self.service_qualifier.service,
                description=description,
                read_quota_bytes=self.service_read_quota,
                write_quota_bytes=self.service_write_quota,
                scan_quota_bytes=self.service_scan_quota,
                update_if_exists=True,
            )
            self.logger.info("Staged service '%s'", self.service_qualifier)
        else:
            admin_data = await self.client.admin.services.find_by_organization_project_and_name(
                organization_name=self.service_qualifier.organization,
                project_name=self.service_qualifier.project,
                service_name=self.service_qualifier.service,
            )
            self.logger.info("Found service '%s'", self.service_qualifier)
        self.service_id = uuid.UUID(hex=admin_data["serviceID"])

    async def _teardown_service(self):
        try:
            await self._stage_service(create=False)
            await self.client.admin.services.delete(self.service_id)
            self.logger.info("Deleted service '%s'", self.service_qualifier)
        except GraphQLError as e:
            if " not found" in str(e):
                return
            raise

    async def _stage_checkpointer(self, create: bool):
        metastream_name = self.service_qualifier.service + "-checkpoints"
        self.checkpoints = await self.client.checkpointer(
            project_path=f"{self.service_qualifier.organization}/{self.service_qualifier.project}",
            metastream_name=metastream_name,
            create_metastream=create,
            key_prefix=f"{self.version}:",
        )
        if create:
            self.logger.info(
                "Staged checkpointer '%s'", self.checkpoints.instance.stream._qualifier
            )
        else:
            self.logger.info("Found checkpointer '%s'", self.checkpoints.instance.stream._qualifier)

    async def _teardown_checkpointer(self):
        try:
            await self._stage_checkpointer(create=False)
            await self.checkpoints.instance.stream.delete()
            self.logger.info(
                "Deleted checkpointer '%s'", self.checkpoints.instance.stream._qualifier
            )
        except GraphQLError as e:
            if " not found" in str(e):
                return
            raise

    def _add_input_stream(self, stream_path: str):
        qualifier = StreamQualifier.from_path(stream_path)
        self._input_qualifiers.add(qualifier)
        return qualifier

    async def _stage_input_stream(self, qualifier: StreamQualifier):
        if qualifier in self._output_qualifiers:
            raise ValueError(f"Cannot use stream '{qualifier}' as both input and output")
        subscription_path = f"{self.service_qualifier.organization}/"
        subscription_path += f"{self.service_qualifier.project}/"
        subscription_path += self.service_qualifier.service
        consumer = await self.client.consumer(
            stream_path=str(qualifier),
            subscription_path=subscription_path,
            checkpointer=self.checkpoints,
        )
        self._inputs[qualifier] = consumer
        self.logger.info(
            "Using input stream '%s' (version: %d)", qualifier, consumer.instance.version
        )

    def _add_output_stream(
        self,
        stream_path: str,
        **kwargs,
    ) -> StreamQualifier:
        # use service organization and project if "stream_path" is just a name
        if "/" not in stream_path:
            stream_path = (
                f"{self.service_qualifier.organization}/"
                + f"{self.service_qualifier.project}/"
                + stream_path
            )
        qualifier = StreamQualifier.from_path(stream_path)
        self._output_qualifiers.add(qualifier)
        self._output_kwargs[qualifier] = kwargs
        return qualifier

    async def _stage_output_stream(self, qualifier: StreamQualifier, create: bool):
        kwargs = self._output_kwargs[qualifier]
        create = create and "schema" in kwargs
        if create:
            stream = await self.client.create_stream(
                stream_path=str(qualifier),
                update_if_exists=True,
                **kwargs,
            )
        else:
            stream = await self.client.find_stream(stream_path=str(qualifier))

        instance = await stream.create_instance(version=self.version, update_if_exists=True)
        self._outputs[qualifier] = instance

        if create:
            self.logger.info(
                "Staged output stream '%s' (using version %s)", qualifier, self.version
            )
        else:
            self.logger.info("Found output stream '%s' (using version %s)", qualifier, self.version)

    async def _teardown_output_stream(self, qualifier: StreamQualifier):
        try:
            kwargs = self._output_kwargs[qualifier]
            if "schema" not in kwargs:
                return
            await self._stage_output_stream(qualifier=qualifier, create=False)
            instance = self._outputs[qualifier]
            await instance.stream.delete()
            self.logger.info("Deleted output '%s'", qualifier)
        except GraphQLError as e:
            if " not found" in str(e):
                return
            raise

    async def _update_service_permissions(
        self,
        stream: Stream,
        read: bool = None,
        write: bool = None,
    ):
        await self.client.admin.services.update_permissions_for_stream(
            service_id=str(self.service_id),
            stream_id=str(stream.stream_id),
            read=read,
            write=write,
        )
