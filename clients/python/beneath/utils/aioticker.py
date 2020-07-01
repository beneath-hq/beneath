import asyncio
import time

class AIOTicker:

  def __init__(self, at_least_every_ms: int, at_most_every_ms: int):
    self._at_most_every = at_most_every_ms / 1000
    self._at_least_every = at_least_every_ms / 1000

    self._last_tick = 0
    self._wait_task: asyncio.Task = None
    self._triggered = False
    self._stop = False

  def trigger(self):
    self._triggered = True
    if self._wait_task:
      self._wait_task.cancel()

  def stop(self):
    self._stop = True
    self.trigger()

  def __aiter__(self):
    return self

  async def __anext__(self):
    if not self._triggered:
      try:
        self._wait_task = asyncio.create_task(asyncio.sleep(self._at_least_every))
        await self._wait_task
        self._wait_task = None
        self._triggered = True
      except asyncio.CancelledError:
        self._wait_task = None
        assert self._triggered

    if self._stop:
      raise StopAsyncIteration()

    delta = time.time() - self._last_tick
    if delta < self._at_most_every:
      await asyncio.sleep(self._at_most_every - delta)

    self._last_tick = time.time()
    self._triggered = False
    return
