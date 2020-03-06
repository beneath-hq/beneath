import asyncio
from typing import Awaitable, Callable, TypeVar

Item = TypeVar('Item')
ProcessFunc = Callable[[Item], Awaitable[None]]


class AIOWorkerPool:

  def __init__(self, callback: ProcessFunc, maxsize=0):
    self._queue = asyncio.Queue(maxsize=maxsize)
    self._callback = callback
    self._run_task: asyncio.Task = None

  async def enqueue(self, item: Item):
    await self._queue.put(item)

  async def join(self):
    if not self._run_task:
      raise ValueError("you cannot call `join` before calling `run`")
    await self._queue.join()
    self._run_task.cancel()
    try:
      await self._run_task
    except asyncio.CancelledError:
      pass


  async def run(self, workers: int):
    tasks = (self._work() for _ in range(workers))
    self._run_task = asyncio.gather(*tasks)
    try:
      await self._run_task
    except asyncio.CancelledError:
      pass

  async def _work(self):
    while True:
      item = await self._queue.get()
      await self._callback(item)
      self._queue.task_done()
