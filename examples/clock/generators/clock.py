import asyncio
from datetime import datetime, timedelta
from config import START_TIME, FREQUENCY

start = datetime.fromisoformat(START_TIME)
tz = start.tzinfo
delta = timedelta(seconds=float(FREQUENCY))


async def ticker(p):
    last_tick = await p.get_checkpoint("time", default=start)

    while True:
        next_tick = last_tick + delta
        now = datetime.now(tz=tz)
        if next_tick >= now:
            await asyncio.sleep((next_tick - now).total_seconds())

        tick = last_tick + delta
        yield {"time": tick}
        await p.set_checkpoint("time", tick.timestamp())
        last_tick = tick
