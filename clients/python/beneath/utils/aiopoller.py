import asyncio
import time
from typing import Awaitable, Callable

PollFunc = Callable[[], Awaitable[None]]


class AIOPoller:

  def __init__(self, poll: PollFunc, at_least_every: float, at_most_every: float):
    self._poll = poll
    self._at_most_every = at_most_every
    self._at_least_every = at_least_every

    self._last_poll = 0
    self._wait_task: asyncio.Task = None
    self._triggered = False

  def trigger(self):
    self._triggered = True
    if self._wait_task:
      self._wait_task.cancel()

  async def run(self):
    while True:
      # wait if not already triggered
      if not self._triggered:
        try:
          self._wait_task = asyncio.create_task(asyncio.sleep(self._at_least_every))
          await self._wait_task
          self._wait_task = None
          self._triggered = True
        except asyncio.CancelledError:
          self._wait_task = None
          assert self._triggered

      # wait if we at_most_every seconds hasn't passed since last poll
      delta = time.time() - self._last_poll
      if delta < self._at_most_every:
        await asyncio.sleep(self._at_most_every - delta)

      # run poll
      self._last_poll = time.time()
      self._triggered = False
      await self._poll()
