import asyncio
import logging
from typing import Generic, TypeVar

BufferValue = TypeVar('BufferValue')


class AIODelayBuffer(Generic[BufferValue]):
  """
  AIODelayBuffer provides an asyncio-based means of buffering and flushing data based on
  the size of buffered values or time passed since the first value was written to the buffer.

  We use it to buffer writes for `max_delay_ms` before sending them in a single batched request
  over the network (with forced buffer flushes at `max_size`).

  It only lets one buffer be open at any moment. If a write is attempted when the buffer is full
  or flushing, the write will not return until the flush of the previous buffer has completed.

  This class is NOT thread-safe.
  """

  def __init__(self, max_delay_ms: int, max_size: int, max_count: int):
    self._max_delay = max_delay_ms / 1000
    self._max_size = max_size
    self._max_count = max_count

    self._delay_task: asyncio.Task = None
    self._delayed_flush_task: asyncio.Task = None

    self._started = False
    self._flushing = False
    self._buffer_size = 0
    self._buffer_count = 0
    self._reset()

  # PROPERTIES

  @property
  def running(self):
    return self._started

  # OVERRIDES

  # pylint: disable=no-self-use
  def _reset(self):
    raise Exception("AIODelayBuffer subclasses must implement _reset")

  # pylint: disable=no-self-use
  def _merge(self, value: BufferValue):
    raise Exception("AIODelayBuffer subclasses must implement _merge")

  # pylint: disable=no-self-use
  async def _flush(self):
    raise Exception("AIODelayBuffer subclasses must implement _flush")

  # LIFECYCLE

  async def __aenter__(self):
    await self.start()
    return self

  async def __aexit__(self, exc_type, exc, tb):
    if not exc_type:
      await self.stop()

  async def start(self):
    if self._started:
      raise Exception("Already called start")
    self._started = True

  async def stop(self):
    await self.force_flush()
    self._started = False

  async def write(self, value: BufferValue, size: int) -> asyncio.Task:
    """
    Adds value to the buffer. When an awaited call to write returns, the value has been added to the buffer,
    but not been flushed yet. If you wish to wait until the write has been flushed, await the task returned
    by write. For example,

        task = await buffer.write(value=..., size=..)
        await task
    """

    # check value can ever fit in buffer
    if size > self._max_size:
      raise ValueError(f"Value exceeds maximum size (size={size} max_size={self._max_size} value={value})")

    # trigger and/or wait for flush if a) a flush is in progress, or b) value would cause size overflow
    loops = 0
    while self._flushing or (self._buffer_size + size > self._max_size) or (self._buffer_count == self._max_count):
      assert self._delayed_flush_task is not None
      await self.force_flush()
      loops += 1
      if loops > 5:
        logging.warning("Unfortunate scheduling blocked write to buffer %i times (try to limit concurrent writes)", loops)

    # now we know we're not flushing and the value fits; and execution will not be "interrupted" until next "await"

    # add to buffer
    self._merge(value)
    self._buffer_size += size
    self._buffer_count += 1

    # if a delayed flush isn't already scheduled for this batch, schedule it now
    if not self._delayed_flush_task:
      self._delay_task = asyncio.create_task(asyncio.sleep(self._max_delay))
      self._delayed_flush_task = asyncio.create_task(self._delayed_flush())
      self._delayed_flush_task.add_done_callback(self._delayed_flush_done)

    return self._delayed_flush_task

  async def force_flush(self):
    if self._delay_task:
      self._delay_task.cancel()
    if self._delayed_flush_task:
      await self._delayed_flush_task

  async def _delayed_flush(self):
    # wait for delay
    if self._delay_task:
      try:
        await self._delay_task
      except asyncio.CancelledError:
        pass
    # flush
    self._flushing = True
    await self._flush()
    self._reset()
    self._flushing = False
    self._buffer_size = 0
    self._buffer_count = 0
    self._delay_task = None
    self._delayed_flush_task = None

  def _delayed_flush_done(self, task: asyncio.Task):
    if task.exception():
      logging.error("Error in buffer flush background loop", exc_info=task.exception())
