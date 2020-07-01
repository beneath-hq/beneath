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
import sys
from typing import Any, Dict, List, Set, Tuple
import uuid

import msgpack

from beneath import config
from beneath.client import Client
from beneath.instance import StreamInstance
from beneath.logging import logger
from beneath.stream import Stream
from beneath.utils import AIODelayBuffer, ServiceQualifier, StreamQualifier
from beneath.writer import Writer


class Action(Enum):
  test = "test"
  stage = "stage"
  launch = "launch"


class Strategy(Enum):
  continuous = "continuous"
  delta = "delta"
  batch = "batch"


SERVICE_STATE_NAME = "services-state"
SERVICE_STATE_LOG_RETENTION = timedelta(hours=6)
SERVICE_STATE_SCHEMA = """
  type ServiceState @stream @key(fields: ["service_id", "version", "key"]) {
    service_id: Bytes16!
    version: Int!
    key: String!
    value: Bytes
  }
"""


class BasePipeline:

  instances: Dict[StreamQualifier, StreamInstance]
  state_instance: StreamInstance
  writer: Writer
  state_writer: BasePipeline._StateWriter
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
    client: Client = None,
    write_delay_ms: int = config.DEFAULT_WRITE_DELAY_MS,
    write_state_delay_ms: int = config.DEFAULT_WRITE_PIPELINE_STATE_DELAY_MS,
  ):
    self.action = self._arg_or_env("action", action, "BENEATH_PIPELINE_ACTION", fn=Action)
    self.strategy = self._arg_or_env("strategy", strategy, "BENEATH_PIPELINE_STRATEGY", fn=Strategy)
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
    self.client = client if client else Client()
    self.write_delay_ms = write_delay_ms
    self.write_state_delay_ms = write_state_delay_ms
    self.instances: Dict[StreamQualifier, StreamInstance] = {}
    self.state_instance = None
    self.writer = None
    self.state_writer = None
    self.service_id = None
    self.description = None
    self._stage_tasks = []
    self._input_qualifiers = set()
    self._output_qualifiers = set()
    self.logger = logger
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

  def _init(self):
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
    loop.run_until_complete(self.run())

  async def run(self):
    await self.stage()
    if self.action == Action.stage:
      return

    dry = self.action == Action.test
    if dry:
      logger.info("Running in TEST mode: Records will be printed, but NOT saved to Beneath")
    else:
      logger.info("Running in PRODUCTION mode: Records will be saved to Beneath")

    self.writer = self.client.writer(dry=dry, write_delay_ms=self.write_delay_ms)
    self.state_writer = self._StateWriter(self)

    await self._before_run()
    await self.writer.start()
    await self.state_writer.start()
    try:
      await self._run()
      await self._after_run()
    finally:
      await self.state_writer.stop()
      await self.writer.stop()

  # SERVICE STATE

  class _StateWriter(AIODelayBuffer[Tuple[str, Any]]):

    state: Dict[str, Any]

    def __init__(self, pipeline: BasePipeline):
      super().__init__(max_delay_ms=pipeline.write_state_delay_ms, max_size=sys.maxsize)
      self.pipeline = pipeline

    def _reset(self):
      self.state = {}

    def _merge(self, value: Tuple[str, Any]):
      (key, state) = value
      self.state[key] = state

    async def _flush(self):
      records = ({
        "service_id": self.pipeline.service_id.bytes,
        "version": self.pipeline.version,
        "key": key,
        "value": msgpack.packb(state),
      } for (key, state) in self.state.items())
      await self.pipeline.writer.write(instance=self.pipeline.state_instance, records=records)

    # pylint: disable=arguments-differ
    async def write(self, key: str, state: Any):
      await super().write(value=(key, state), size=0)

  async def get_state(self, key: str, default: Any = None) -> Any:
    cursor = await self.state_instance.query_index(
      filter=json.dumps({
        "service_id": "0x" + self.service_id.bytes.hex(), # not pretty
        "version": self.version,
        "key": key,
      })
    )

    value = default
    batch = await cursor.read_next(limit=1)
    if batch is not None:
      records = list(batch)
      if len(records) > 0:
        value = records[0]["value"]
        value = msgpack.unpackb(value)

    return value

  async def set_state(self, key: str, value: Any):
    await self.state_writer.write(key, value)

  # STAGING

  async def stage(self):
    logger.info("Staging service '%s'", self.service_qualifier)
    await self._stage_service()
    await asyncio.gather(self._stage_service_state(), *self._stage_tasks)
    logger.info("Finished staging pipeline for service '%s'", self.service_qualifier)

  async def _stage_service(self):
    admin_data = await self.client.admin.services.stage(
      organization_name=self.service_qualifier.organization,
      project_name=self.service_qualifier.project,
      service_name=self.service_qualifier.service,
      description=self.description,
      read_quota_bytes=self.service_read_quota,
      write_quota_bytes=self.service_write_quota,
    )
    self.service_id = uuid.UUID(hex=admin_data["serviceID"])

  async def _stage_service_state(self):
    qualifier = StreamQualifier(
      organization=self.service_qualifier.organization,
      project=self.service_qualifier.project,
      stream=SERVICE_STATE_NAME,
    )
    stream = await self.client.stage_stream(
      stream_path=str(qualifier),
      schema=SERVICE_STATE_SCHEMA,
      log_retention=SERVICE_STATE_LOG_RETENTION,
      use_warehouse=False,
    )
    await self._update_service_permissions(stream, read=True, write=True)
    instance = await stream.stage_instance(version=0, make_primary=True)
    self.state_instance = instance
    logger.info(
      "Staged meta-stream for pipeline state '%s' (using primary instance %s)", qualifier, instance.instance_id
    )

  def _add_input_stream(self, stream_path: str) -> StreamQualifier:
    qualifier = StreamQualifier.from_path(stream_path)
    if (qualifier in self._input_qualifiers) or (qualifier in self._output_qualifiers):
      raise ValueError(f"Stream {qualifier} already added as input or output")
    self._input_qualifiers.add(qualifier)
    self._stage_tasks.append(self._stage_input_stream(stream_path))
    return qualifier

  async def _stage_input_stream(self, stream_path: str):
    stream = await self.client.find_stream(stream_path=stream_path)
    instance = stream.primary_instance
    if not instance:
      raise ValueError(f"Input stream {stream_path} doesn't have a primary instance")
    assert stream.qualifier not in self.instances
    await self._update_service_permissions(stream, read=True)
    self.instances[stream.qualifier] = stream.primary_instance
    logger.info("Staged input stream '%s' (using primary instance %s)", stream_path, instance.instance_id)

  def _add_output_stream(
    self,
    stream_path: str,
    schema: str = None,
    use_index: bool = None,
    use_warehouse: bool = None,
    retention: timedelta = None,
    log_retention: timedelta = None,
    index_retention: timedelta = None,
    warehouse_retention: timedelta = None,
  ) -> StreamQualifier:
    qualifier = StreamQualifier.from_path(stream_path)
    if qualifier in self._output_qualifiers:
      return qualifier
    if qualifier in self._input_qualifiers:
      raise ValueError(f"Cannot use stream {qualifier} added as both input and output")
    self._output_qualifiers.add(qualifier)
    self._stage_tasks.append(
      self._stage_output_stream(
        stream_path=stream_path,
        schema=schema,
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
    use_index: bool = None,
    use_warehouse: bool = None,
    retention: timedelta = None,
    log_retention: timedelta = None,
    index_retention: timedelta = None,
    warehouse_retention: timedelta = None,
  ):
    if not schema:
      stream = await self.client.find_stream(stream_path)
    else:
      stream = await self.client.stage_stream(
        stream_path=stream_path,
        schema=schema,
        use_index=use_index,
        use_warehouse=use_warehouse,
        retention=retention,
        log_retention=log_retention,
        index_retention=index_retention,
        warehouse_retention=warehouse_retention,
      )
    assert stream.qualifier not in self.instances
    instance = await stream.stage_instance(version=self.version)
    await self._update_service_permissions(stream, write=True)
    self.instances[stream.qualifier] = instance
    logger.info("Staged output stream '%s' (using instance %s)", stream_path, instance.instance_id)

  async def _update_service_permissions(self, stream: Stream, read: bool = None, write: bool = None):
    await self.client.admin.services.update_permissions_for_stream(
      service_id=self.service_id,
      stream_id=stream.stream_id,
      read=read,
      write=write,
    )
