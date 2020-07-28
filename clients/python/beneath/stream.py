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
from typing import Iterable, Tuple
import uuid

from fastavro import parse_schema
from fastavro import schemaless_reader
from fastavro import schemaless_writer

from beneath.instance import DryStreamInstance, StreamInstance
from beneath.proto import gateway_pb2
from beneath.utils import (
  ms_to_datetime,
  ms_to_pd_timestamp,
  StreamQualifier,
  timestamp_to_ms,
)


class Stream:
  """
  Represents a data-plane connection to a stream. I.e., this class is used to
  query, peek, consume and write to a stream. You cannot use this class to do control-plane actions
  like creating streams or updating their details (use `client.admin.streams` directly).
  """

  client: Client
  qualifier: StreamQualifier
  admin_data: dict
  stream_id: uuid.UUID
  avro_schema_parsed: dict
  primary_instance: StreamInstance

  # INITIALIZATION

  def __init__(self):
    self.client: Client = None
    self.qualifier: StreamQualifier = None
    self.admin_data: dict = None
    self.stream_id: uuid.UUID = None
    self.avro_schema_parsed: dict = None
    self.primary_instance: StreamInstance = None

  @classmethod
  async def make(cls, client: Client, qualifier: StreamQualifier, admin_data=None) -> Stream:
    stream = Stream()
    stream.client = client
    stream.qualifier = qualifier

    if not admin_data:
      # pylint: disable=protected-access
      admin_data = await stream._load_admin_data()
    # pylint: disable=protected-access
    stream._set_admin_data(admin_data)

    return stream

  async def _load_admin_data(self):
    return await self.client.admin.streams.find_by_organization_project_and_name(
      organization_name=self.qualifier.organization,
      project_name=self.qualifier.project,
      stream_name=self.qualifier.stream,
    )

  def _set_admin_data(self, admin_data):
    self.admin_data = admin_data
    self.stream_id = uuid.UUID(hex=self.admin_data["streamID"])
    self.avro_schema_parsed = parse_schema(json.loads(self.admin_data["avroSchema"]))
    if "primaryStreamInstance" in self.admin_data and self.admin_data["primaryStreamInstance"] is not None:
      self.primary_instance = StreamInstance(stream=self, admin_data=self.admin_data["primaryStreamInstance"])

  # MANAGEMENT

  async def delete(self):
    await self.client.admin.streams.delete(self.stream_id)

  # INSTANCES

  async def find_instances(self) -> Iterable[StreamInstance]:
    instances = await self.client.admin.streams.find_instances(self.admin_data["streamID"])
    instances = [StreamInstance(stream=self, admin_data=i) for i in instances]
    return instances

  async def stage_instance(self, version: int, make_primary=None, make_final=None, dry=False) -> StreamInstance:
    if dry:
      return DryStreamInstance(self, version=version, primary=make_primary, final=make_final)
    instance = await self.client.admin.streams.stage_instance(
      stream_id=self.admin_data["streamID"],
      version=version,
      make_final=make_final,
      make_primary=make_primary,
    )
    instance = StreamInstance(stream=self, admin_data=instance)
    return instance

  # ENCODING / DECODING RECORDS

  def record_to_pb(self, record: Mapping) -> Tuple[gateway_pb2.Record, int]:
    if not isinstance(record, Mapping):
      raise TypeError("write error: record must be a mapping, got {}".format(record))
    avro = self._encode_avro(record)
    timestamp = self._extract_record_timestamp(record)
    pb = gateway_pb2.Record(avro_data=avro, timestamp=timestamp)
    return (pb, pb.ByteSize())

  def pb_to_record(self, pb: gateway_pb2.Record, to_dataframe: bool) -> Mapping:
    record = self._decode_avro(pb.avro_data)
    record["@meta.timestamp"] = ms_to_pd_timestamp(pb.timestamp) if to_dataframe else ms_to_datetime(pb.timestamp)
    return record

  def _encode_avro(self, record: Mapping):
    writer = io.BytesIO()
    schemaless_writer(writer, self.avro_schema_parsed, record)
    result = writer.getvalue()
    writer.close()
    return result

  def _decode_avro(self, data):
    reader = io.BytesIO(data)
    record = schemaless_reader(reader, self.avro_schema_parsed)
    reader.close()
    return record

  @staticmethod
  def _extract_record_timestamp(record: Mapping) -> int:
    if ("@meta" in record) and ("timestamp" in record["@meta"]):
      return timestamp_to_ms(record["@meta"]["timestamp"])
    return 0  # 0 tells the server to set timestamp to its current time
