import asyncio
from typing import Awaitable, Callable, List, Iterable, Tuple, TypeVar

BufferValue = TypeVar('BufferValue')
FlushResult = TypeVar('FlushResult')
FlushFunc = Callable[[List[BufferValue]], Awaitable[FlushResult]]


class AIOWindowedBuffer:
  """
  AIOWindowedBuffer provides an asyncio-based means of buffering and flushing data based on
  number of buffered values, size of buffered values, and time passed since the first value
  was written to the buffer.

  We use it to buffer writes for `delay` seconds before sending them in a single batched request
  over the network (with forced buffer flushes at `max_len` or `max_bytes`). A write will not return
  until the underlying buffer has been flushed.

  This class is NOT thread-safe.
  """

  # The ordering of creating tasks, awaiting tasks, and resetting the buffer (_expire_buffer)
  # in this class is not very intuitive. That's due to the fact that whenever we `await`,
  # we hand back control to the event loop, and then cannot make any assumptions about what gets
  # called when. Conversely, we're guaranteed uninterrupted execution between `await`s.

  def __init__(self, flush: FlushFunc, delay_seconds: float, max_len: int, max_bytes: int):
    # settings
    self._delay = delay_seconds
    self._max_len = max_len
    self._max_bytes = max_bytes
    self._flush_func = flush
    # state
    self._buffer: List[BufferValue] = []
    self._buffer_bytes = 0
    self._delayed_expire_task: asyncio.Task = None
    self._delayed_flush_task: asyncio.Task = None

  async def write_one(self, value: BufferValue, nbytes: int, force_flush=False) -> FlushResult:
    task = self._write(value, nbytes)
    if force_flush:
      self._force_flush() # discarding result because it's equal to `task`
    res = await task
    return res

  async def write_many(self, values: Iterable[Tuple[BufferValue, int]], force_flush=False):
    current = None
    tasks = []
    for (value, nbytes) in values:
      task = self._write(value, nbytes)
      if task is not current:
        current = task
        tasks.append(task)
    if force_flush:
      task = self._force_flush()
      if task is not None and task is not current:
        # assert len(values) == 0 and force_flush = True
        tasks.append(task)
    await asyncio.gather(*tasks)
    return

  def _write(self, value: BufferValue, nbytes: int) -> Awaitable[FlushResult]:
    # check value can ever fit in buffer
    if nbytes > self._max_bytes:
      raise ValueError(f"value exceeds maximum of {self._max_bytes} bytes")

    # reset buffer if value would cause size overflow
    if self._buffer_bytes + nbytes > self._max_bytes:
      self._delayed_expire_task.cancel()
      self._expire_buffer()

    # add to buffer
    self._buffer.append(value)
    self._buffer_bytes += nbytes

    # if a delayed flush isn't already scheduled for this batch, schedule it now
    if not self._delayed_flush_task or self._delayed_flush_task.cancelled():
      self._delayed_expire_task = asyncio.create_task(self._delayed_expire())
      self._delayed_flush_task = asyncio.create_task(self._delayed_flush(self._buffer, self._delayed_expire_task))

    # trigger flush immediately if buffer is full
    if len(self._buffer) == self._max_len:
      return self._force_flush()

    return self._delayed_flush_task

  def _force_flush(self) -> Awaitable[FlushResult]:
    if self._delayed_flush_task is not None:
      flush_task = self._delayed_flush_task  # get pointer to self._delayed_flush_task before we reset self
      self._delayed_expire_task.cancel()
      self._expire_buffer()
      return flush_task
    return None

  def _expire_buffer(self):
    # it's safe to directly reset these values since _delayed_flush holds pointers
    # to the old values (creating a sort of implicit buffer of buffers to flush)
    self._buffer: List[BufferValue] = []
    self._buffer_bytes = 0
    self._delayed_expire_task: asyncio.Task = None
    self._delayed_flush_task: asyncio.Task = None

  async def _delayed_expire(self):
    await asyncio.sleep(self._delay)
    self._expire_buffer()  # not called if _delayed_expire_task is cancelled

  async def _delayed_flush(self, buffer: List[BufferValue], delayed_expire_task: asyncio.Task) -> FlushResult:
    try:
      await delayed_expire_task
    except asyncio.CancelledError:
      pass
    return await self._flush_func(buffer)
