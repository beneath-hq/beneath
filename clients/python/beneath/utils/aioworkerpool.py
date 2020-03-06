import asyncio
from typing import Awaitable, Callable, TypeVar

Item = TypeVar('Item')
ProcessFunc = Callable[[Item], Awaitable[None]]


class AIOWorkerPool:

  def __init__(self, callback: ProcessFunc, maxsize=0):
    self._queue = asyncio.Queue(maxsize=maxsize)
    self._callback = callback

  async def enqueue(self, item: Item):
    await self._queue.put(item)

  async def run(self, workers: int):
    tasks = (self._work() for _ in range(workers))
    return await asyncio.gather(*tasks)

  async def _work(self):
    while True:
      item = await self._queue.get()
      print(f"SIZE: {self._queue.qsize()}")
      await self._callback(item)
      self._queue.task_done()
