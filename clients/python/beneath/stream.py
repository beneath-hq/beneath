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
from typing import Iterable, List, Tuple, Union
import uuid

from fastavro import parse_schema
from fastavro import schemaless_reader
from fastavro import schemaless_writer
import pandas as pd

from beneath.config import DEFAULT_READ_BATCH_SIZE, DEFAULT_WRITE_DELAY_SECONDS, DEFAULT_WRITE_BATCH_SIZE, DEFAULT_WRITE_BATCH_BYTES
from beneath.proto import gateway_pb2
from beneath.utils import AIOWindowedBuffer, ms_to_datetime, ms_to_pd_timestamp, timestamp_to_ms


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
    max_write_delay_seconds: float = DEFAULT_WRITE_DELAY_SECONDS
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
      max_len=DEFAULT_WRITE_BATCH_SIZE,
      max_bytes=DEFAULT_WRITE_BATCH_BYTES,
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

  async def write(self, records: Union[Iterable[Mapping], Mapping], immediate=False):
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

  # READING RECORDS

  async def query(self, where: str = None) -> Cursor:
    resp = await self.client.connection.query(instance_id=self.instance_id, where=where)
    assert len(resp.replay_cursors) <= 1 and len(resp.change_cursors) <= 1
    replay = resp.replay_cursors[0] if len(resp.replay_cursors) > 0 else None
    changes = resp.change_cursors[0] if len(resp.change_cursors) > 0 else None
    return Cursor(stream=self, instance_id=self.instance_id, replay_cursor=replay, changes_cursor=changes)

  async def peek(self) -> Cursor:
    resp = await self.client.connection.peek(instance_id=self.instance_id)
    return Cursor(stream=self, instance_id=self.instance_id, replay_cursor=resp.rewind_cursor, changes_cursor=resp.changes_cursor)

  # EASY HELPERS

  def easy_read(self, to_dataframe=True):
    pass

  def process(self, record_cb, commit_strategy):
    pass

class Cursor:

  def __init__(self, stream: Stream, instance_id: uuid.UUID, replay_cursor: bytes, changes_cursor: bytes):
    self.stream = stream
    self.instance_id = instance_id
    self.replay_cursor = replay_cursor
    self.changes_cursor = changes_cursor

  @property
  def _columns(self):
    # pylint: disable=protected-access
    return [field["name"] for field in self.stream._avro_schema_parsed["fields"]].append("@meta.timestamp")

  async def fetch_next(self, limit: int = DEFAULT_READ_BATCH_SIZE, to_dataframe=False) -> Iterable[Mapping]:
    if not self.replay_cursor:
      return None
    resp = await self.stream.client.connection.read(instance_id=self.instance_id, cursor=self.replay_cursor, limit=limit)
    self.replay_cursor = resp.next_cursor
    return self._parse_pbs(pbs=resp.records, to_dataframe=to_dataframe)

  def _parse_pbs(self, pbs: List[gateway_pb2.Record], to_dataframe: bool) -> Iterable[Mapping]:
    records = (self._pb_to_record(pb, to_dataframe) for pb in pbs)
    if to_dataframe:
      return pd.DataFrame(records, columns=self._columns)
    return records

  def _pb_to_record(self, pb: gateway_pb2.Record, to_dataframe: bool) -> Mapping:
    record = self._decode_avro(pb.avro_data)
    record["@meta.timestamp"] = ms_to_pd_timestamp(pb.timestamp) if to_dataframe else ms_to_datetime(pb.timestamp)
    return record

  def _decode_avro(self, data):
    reader = io.BytesIO(data)
    # pylint: disable=protected-access
    record = schemaless_reader(reader, self.stream._avro_schema_parsed)
    reader.close()
    return record

  # async def fetch_all(self, max_rows=None, max_bytes=None, warn_max=True, to_dataframe=False) -> Iterable[Mapping]:
  #   pass


  # fetch_changes
  # subscribe_changes
