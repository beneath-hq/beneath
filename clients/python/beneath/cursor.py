# Allows us to use StreamInstance as a type hint without an import cycle
# pylint: disable=wrong-import-position,ungrouped-imports
from __future__ import annotations
from typing import TYPE_CHECKING
if TYPE_CHECKING:
  from beneath.instance import StreamInstance

import asyncio
from collections.abc import Mapping
import sys
from typing import AsyncIterator, Awaitable, Callable, Iterable, List
import warnings

import pandas as pd

from beneath import config
from beneath.utils import AIOTicker


class Cursor:

  def __init__(self, instance: StreamInstance, replay_cursor: bytes, changes_cursor: bytes):
    self.instance = instance
    self.replay_cursor = replay_cursor
    self.changes_cursor = changes_cursor

  @property
  def _top_level_columns(self):
    return [field["name"] for field in self.instance.stream.avro_schema_parsed["fields"]].append("@meta.timestamp")

  async def read_next(self, limit: int = config.DEFAULT_READ_BATCH_SIZE, to_dataframe=False) -> Iterable[Mapping]:
    batch = await self._read_next_replay(limit=limit)
    if batch is None:
      return None
    records = (self.instance.stream.pb_to_record(pb, to_dataframe) for pb in batch)
    if to_dataframe:
      return pd.DataFrame(records, columns=self._top_level_columns)
    return records

  async def read_next_changes(
    self,
    limit: int = config.DEFAULT_READ_BATCH_SIZE,
    to_dataframe=False,
  ) -> Iterable[Mapping]:
    batch = await self._read_next_changes(limit=limit)
    if batch is None:
      return None
    records = (self.instance.stream.pb_to_record(pb, to_dataframe) for pb in batch)
    if to_dataframe:
      return pd.DataFrame(records, columns=self._top_level_columns)
    return records

  async def read_all(
    self,
    max_records=None,
    max_bytes=config.DEFAULT_READ_ALL_MAX_BYTES,
    batch_size=config.DEFAULT_READ_BATCH_SIZE,
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
        record = self.instance.stream.pb_to_record(pb=pb, to_dataframe=False)
        records.append(record)
        batch_len += 1
        bytes_loaded += len(pb.avro_data)

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
      return pd.DataFrame(records, columns=self._top_level_columns)
    return records

  async def _read_next_replay(self, limit: int):
    if not self.replay_cursor:
      return None
    resp = await self.instance.stream.client.connection.read(
      cursor=self.replay_cursor,
      limit=limit,
    )
    self.replay_cursor = resp.next_cursor
    return resp.records

  async def _read_next_changes(self, limit: int):
    if not self.changes_cursor:
      return None
    resp = await self.instance.stream.client.connection.read(
      cursor=self.changes_cursor,
      limit=limit,
    )
    self.changes_cursor = resp.next_cursor
    return resp.records

  async def subscribe_changes(
    self,
    batch_size=config.DEFAULT_READ_BATCH_SIZE,
    poll_at_most_every_ms=config.DEFAULT_SUBSCRIBE_POLL_AT_MOST_EVERY_MS,
  ) -> AsyncIterator[List[Mapping]]:
    """ Wraps subscribe_changes_with_callback as an async iterator. Note that yielded values are batches, not individual records. """
    queue = asyncio.Queue()
    done = Exception("DONE")

    async def callback(records, _):
      await queue.put(records)

    def done_callback(task: asyncio.Task):
      if task.exception():
        queue.put_nowait(task.exception())
      else:
        queue.put_nowait(done)

    coro = self.subscribe_changes_with_callback(
      callback,
      batch_size=batch_size,
      poll_at_most_every_ms=poll_at_most_every_ms,
    )
    task = asyncio.create_task(coro)
    task.add_done_callback(done_callback)

    while True:
      item = await queue.get()
      if isinstance(item, Exception):
        if item == done:
          return
        raise item
      yield item
      queue.task_done()

  async def subscribe_changes_with_callback(
    self,
    callback: Callable[[List[Mapping], Cursor], Awaitable[None]],
    batch_size=config.DEFAULT_READ_BATCH_SIZE,
    poll_at_most_every_ms=config.DEFAULT_SUBSCRIBE_POLL_AT_MOST_EVERY_MS,
  ):
    ticker = AIOTicker(
      at_least_every_ms=config.DEFAULT_SUBSCRIBE_POLL_AT_LEAST_EVERY_MS,
      at_most_every_ms=poll_at_most_every_ms,
    )

    async def _poll():
      while True:
        batch = await self.read_next_changes(limit=batch_size)
        batch = list(batch)
        if len(batch) != 0:
          await callback(batch, self)
        if len(batch) < batch_size:
          break

    async def _tick():
      async for _ in ticker:
        await _poll()

    async def _subscribe():
      subscription = await self.instance.stream.client.connection.subscribe(
        cursor=self.changes_cursor,
      )
      async for _ in subscription:
        ticker.trigger()

    await asyncio.gather(_subscribe(), _tick())
