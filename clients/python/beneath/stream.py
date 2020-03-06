# allows us to use Client as a type hint without an import cycle
# see: https://www.stefaanlippens.net/circular-imports-type-hints-python.html
# pylint: disable=wrong-import-position,ungrouped-imports
from __future__ import annotations
from typing import TYPE_CHECKING
if TYPE_CHECKING:
  from beneath.client import Client

import asyncio
from collections.abc import Mapping
import io
import json
import sys
from typing import Awaitable, Callable, Iterable, List, Tuple, Union
import uuid
import warnings

from fastavro import parse_schema
from fastavro import schemaless_reader
from fastavro import schemaless_writer
import pandas as pd

from beneath.config import (
  BIGQUERY_PROJECT,
  DEFAULT_READ_ALL_MAX_BYTES,
  DEFAULT_READ_BATCH_SIZE,
  DEFAULT_WRITE_DELAY_SECONDS,
  DEFAULT_WRITE_BATCH_SIZE,
  DEFAULT_WRITE_BATCH_BYTES,
  DEFAULT_SUBSCRIBE_PREFETCHED_RECORDS,
  DEFAULT_SUBSCRIBE_CONCURRENT_CALLBACKS,
  DEFAULT_SUBSCRIBE_POLL_AT_LEAST_EVERY_SECONDS,
  DEFAULT_SUBSCRIBE_POLL_AT_MOST_EVERY_SECONDS,
)
from beneath.proto import gateway_pb2
from beneath.utils import (
  AIOPoller,
  AIOWorkerPool,
  AIOWindowedBuffer,
  ms_to_datetime,
  ms_to_pd_timestamp,
  timestamp_to_ms,
)


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

  async def _ensure_loaded(self):
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
    if not self._loaded:
      await self._ensure_loaded()
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
    if not self._loaded:
      await self._ensure_loaded()
    resp = await self.client.connection.query(instance_id=self.instance_id, where=where)
    assert len(resp.replay_cursors) <= 1 and len(resp.change_cursors) <= 1
    replay = resp.replay_cursors[0] if len(resp.replay_cursors) > 0 else None
    changes = resp.change_cursors[0] if len(resp.change_cursors) > 0 else None
    return Cursor(stream=self, instance_id=self.instance_id, replay_cursor=replay, changes_cursor=changes)

  async def peek(self) -> Cursor:
    if not self._loaded:
      await self._ensure_loaded()
    resp = await self.client.connection.peek(instance_id=self.instance_id)
    return Cursor(
      stream=self,
      instance_id=self.instance_id,
      replay_cursor=resp.rewind_cursor,
      changes_cursor=resp.change_cursor,
    )

  # MISC

  def get_bigquery_table(self, view=True):
    return "{}.{}.{}{}".format(
      BIGQUERY_PROJECT,
      self.info["project_name"].replace("-", "_"),
      self.info["stream_name"].replace("-", "_"),
      "" if view else "_{}".format(self.instance_id.hex[0:8]),
    )


class Cursor:

  def __init__(self, stream: Stream, instance_id: uuid.UUID, replay_cursor: bytes, changes_cursor: bytes):
    self._stream = stream
    self._instance_id = instance_id
    self._replay_cursor = replay_cursor
    self._changes_cursor = changes_cursor

  @property
  def _columns(self):
    # pylint: disable=protected-access
    return [field["name"] for field in self._stream._avro_schema_parsed["fields"]].append("@meta.timestamp")

  async def fetch_next(self, limit: int = DEFAULT_READ_BATCH_SIZE, to_dataframe=False) -> Iterable[Mapping]:
    batch = await self._read_next_replay(limit=limit)
    if batch is None:
      return None
    records = (self._pb_to_record(pb, to_dataframe) for pb in batch)
    if to_dataframe:
      return pd.DataFrame(records, columns=self._columns)
    return records

  async def fetch_next_changes(self, limit: int = DEFAULT_READ_BATCH_SIZE, to_dataframe=False) -> Iterable[Mapping]:
    if not self._changes_cursor:
      return None
    resp = await self._stream.client.connection.read(
      instance_id=self._instance_id,
      cursor=self._changes_cursor,
      limit=limit,
    )
    self._changes_cursor = resp.next_cursor
    records = (self._pb_to_record(pb, to_dataframe) for pb in resp.records)
    if to_dataframe:
      return pd.DataFrame(records, columns=self._columns)
    return records

  async def fetch_all(
    self,
    max_records=None,
    max_bytes=DEFAULT_READ_ALL_MAX_BYTES,
    batch_size=DEFAULT_READ_BATCH_SIZE,
    warn_max=True,
    to_dataframe=False,
  ) -> Iterable[Mapping]:
    # compute limits
    max_records = max_records if max_records else sys.maxsize
    max_bytes = max_bytes if max_bytes else sys.maxsize

    # loop state
    records = []
    complete = False
    bytes_loaded = 0

    # loop until all records fetched or limits reached
    while len(records) < max_records and bytes_loaded < max_bytes:
      limit = min(max_records - len(records), batch_size)
      assert limit >= 0

      batch = await self._read_next_replay(limit=limit)
      if batch is None:
        complete = True
        break

      batch_len = 0
      for pb in batch:
        record = self._pb_to_record(pb=pb, to_dataframe=False)
        records.append(record)
        batch_len += 1
        bytes_loaded += len(pb.avro_data)

      if batch_len < limit:
        complete = True
        break

    if not complete and warn_max:
      # Jupyter doesn't always display warnings, so also print
      if len(records) >= max_records:
        err = f"Stopped loading because result exceeded max_records={max_records}"
        print(err)
        warnings.warn(err)
      elif bytes_loaded >= max_bytes:
        err = f"Stopped loading because download size exceeded max_bytes={max_bytes}"
        print(err)
        warnings.warn(err)

    if to_dataframe:
      return pd.DataFrame(records, columns=self._columns)
    return records

  async def _read_next_replay(self, limit: int):
    if not self._replay_cursor:
      return None
    resp = await self._stream.client.connection.read(
      instance_id=self._instance_id,
      cursor=self._replay_cursor,
      limit=limit,
    )
    self._replay_cursor = resp.next_cursor
    return resp.records

  def _pb_to_record(self, pb: gateway_pb2.Record, to_dataframe: bool) -> Mapping:
    record = self._decode_avro(pb.avro_data)
    record["@meta.timestamp"] = ms_to_pd_timestamp(pb.timestamp) if to_dataframe else ms_to_datetime(pb.timestamp)
    return record

  def _decode_avro(self, data):
    reader = io.BytesIO(data)
    # pylint: disable=protected-access
    record = schemaless_reader(reader, self._stream._avro_schema_parsed)
    reader.close()
    return record

  async def subscribe_replay(
    self,
    callback: Callable[[Mapping], Awaitable[None]],
    max_prefetched_records=DEFAULT_SUBSCRIBE_PREFETCHED_RECORDS,
    max_concurrent_callbacks=DEFAULT_SUBSCRIBE_CONCURRENT_CALLBACKS,
  ):
    if not self._replay_cursor:
      return None

    records_pool = AIOWorkerPool(callback=callback, maxsize=max_prefetched_records)

    async def _fetch():
      while True:
        batch = await self.fetch_next(limit=min(max_prefetched_records, DEFAULT_READ_BATCH_SIZE))
        if batch is None:
          await records_pool.join()
          return
        for record in batch:
          await records_pool.enqueue(record)

    return await asyncio.gather(
      _fetch(),
      records_pool.run(workers=max_concurrent_callbacks),
    )

  async def subscribe_changes(
    self,
    callback: Callable[[Mapping], Awaitable[None]],
    max_prefetched_records=DEFAULT_SUBSCRIBE_PREFETCHED_RECORDS,
    max_concurrent_callbacks=DEFAULT_SUBSCRIBE_CONCURRENT_CALLBACKS,
  ):
    if not self._changes_cursor:
      raise ValueError("cannot subscribe to changes for this query")

    records_pool = AIOWorkerPool(callback=callback, maxsize=max_prefetched_records)

    async def _poll():
      prev_cursor = None
      while prev_cursor != self._changes_cursor:
        prev_cursor = self._changes_cursor
        batch = await self.fetch_next_changes(limit=min(max_prefetched_records, DEFAULT_READ_BATCH_SIZE))
        for record in batch:
          await records_pool.enqueue(record)

    poller = AIOPoller(
      poll=_poll,
      at_least_every=DEFAULT_SUBSCRIBE_POLL_AT_LEAST_EVERY_SECONDS,
      at_most_every=DEFAULT_SUBSCRIBE_POLL_AT_MOST_EVERY_SECONDS,
    )

    async def _subscribe():
      subscription = await self._stream.client.connection.subscribe(
        instance_id=self._instance_id,
        cursor=self._changes_cursor,
      )
      async for _ in subscription:
        poller.trigger()

    return await asyncio.gather(
      _subscribe(),
      poller.run(),
      records_pool.run(workers=max_concurrent_callbacks),
    )
