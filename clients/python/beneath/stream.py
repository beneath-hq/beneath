# allows us to use Client as a type hint without an import cycle
# see: https://www.stefaanlippens.net/circular-imports-type-hints-python.html
# pylint: disable=wrong-import-position,ungrouped-imports
from __future__ import annotations
from typing import TYPE_CHECKING
if TYPE_CHECKING:
  from beneath.client import Client

from collections.abc import Mapping
import io
import json
from typing import List, Sequence, Tuple, Union
import uuid

from fastavro import parse_schema
# from fastavro import schemaless_reader
from fastavro import schemaless_writer

from beneath.proto import gateway_pb2
from beneath.utils import AIOWindowedBuffer, timestamp_to_ms


class Stream:
  """
  Represents a data-plane connection to a stream. I.e., this class is used to
  query, peek, consume and write to a stream. You cannot use this class to do control-plane actions
  like creating streams or updating their details (use `client.admin.streams` directly).
  """

  # INITIALIZATION

  def __init__(
    self,
    client: Client,
    project: str = None,
    stream: str = None,
    stream_id: str = None,
    max_write_delay_seconds: float = 1.0
  ):
    # check stream_id xor (project, stream)
    if bool(stream_id) == bool(project and stream):
      raise ValueError("must provide either stream_id or (project and stream) parameters, but not both")

    # config
    self.client = client
    self._project_name = project
    self._stream_name = stream
    self._stream_id = stream_id

    # prep state to load
    self.info: dict = None
    self.instance_id: uuid.UUID = None
    self._loaded = False
    self._avro_schema_parsed: dict = None

    # other state
    self._buffer = AIOWindowedBuffer(
      flush=self._flush_buffer,
      delay_seconds=max_write_delay_seconds,
      max_len=10000,
      max_bytes=10000000,
    )

  # LOADING STREAM CONTROL DATA

  async def ensure_loaded(self):
    if not self._loaded:
      data = await self._load_from_admin()
      self._set_admin_data(data)
      self._loaded = True

  async def _load_from_admin(self):
    if self._stream_id:
      return await self.client.admin.streams.find_by_id(stream_id=self._stream_id)
    return await self.client.admin.streams.find_by_project_and_name(
      project_name=self._project_name,
      stream_name=self._stream_name,
    )

  def _set_admin_data(self, data):
    self.info = {
      "stream_id": data['streamID'],
      "stream_name": data['name'],
      "project_name": data['project']['name'],
      "schema": data['schema'],
      "avro_schema": data['avroSchema'],
      "stream_indexes": data['streamIndexes'],
      "external": data['external'],
      "batch": data['batch'],
      "manual": data['manual'],
      "retention_seconds": data['retentionSeconds'],
      "current_instance_id": data['currentStreamInstanceID'],
    }
    self.instance_id = uuid.UUID(hex=self.info["current_instance_id"])
    self._avro_schema_parsed = parse_schema(json.loads(self.info['avro_schema']))

  # WRITING RECORDS

  async def write(self, records: Union[Sequence[Mapping], Mapping], immediate=False):
    if isinstance(records, Mapping):
      records = [records]
    batch = (self._record_to_pb(record) for record in records)
    await self._buffer.write_many(batch, force_flush=immediate)

  def _record_to_pb(self, record: Mapping) -> Tuple[gateway_pb2.Record, int]:
    if not isinstance(record, Mapping):
      raise TypeError("write error: record must be a mapping, got {}".format(record))
    avro = self._encode_avro(record)
    timestamp = self._extract_record_timestamp(record)
    pb = gateway_pb2.Record(avro_data=avro, timestamp=timestamp)
    return (pb, pb.ByteSize())

  def _encode_avro(self, record: Mapping):
    writer = io.BytesIO()
    schemaless_writer(writer, self._avro_schema_parsed, record)
    result = writer.getvalue()
    writer.close()
    return result

  @classmethod
  def _extract_record_timestamp(cls, record: Mapping) -> int:
    if ("@meta" in record) and ("timestamp" in record["@meta"]):
      return timestamp_to_ms(record["@meta"]["timestamp"])
    return 0  # 0 tells the server to set timestamp to its current time

  async def _flush_buffer(self, records: List[gateway_pb2.Record]) -> bytes:
    resp = await self.client.connection.write(instance_id=self.instance_id, records=records)
    return resp.write_id

