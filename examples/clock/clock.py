import asyncio
import beneath
import math
from datetime import datetime, timedelta, timezone

SCHEMA = """
    type Clock @schema {
        time: Timestamp! @key
    }
"""


def make_clock(name: str, start: datetime, delta: timedelta):
    async def _ticker(p: beneath.Pipeline):
        last_tick = await p.checkpoints.get(name, default=start)
        while True:
            now = datetime.now(tz=start.tzinfo)
            next_tick = last_tick + delta
            if next_tick >= now:
                wait = next_tick - now
                await asyncio.sleep(wait.total_seconds())
            yield {"time": next_tick}
            await p.checkpoints.set(name, next_tick)
            last_tick = next_tick

    return _ticker


if __name__ == "__main__":
    p = beneath.Pipeline(parse_args=True)
    p.description = "Pipeline that generates timestamps at fixed intervals"

    start = datetime(year=2021, month=1, day=1, tzinfo=timezone.utc)

    name = "1m"
    clock = p.generate(make_clock(name, start, timedelta(minutes=1)))
    p.write_stream(
        clock,
        stream_path=f"clock-{name}",
        schema=SCHEMA,
        description="Clock that ticks every minute",
    )

    name = "1h"
    clock = p.generate(make_clock(name, start, timedelta(hours=1)))
    p.write_stream(
        clock,
        stream_path=f"clock-{name}",
        schema=SCHEMA,
        description="Clock that ticks every hour",
    )

    name = "1d"
    clock = p.generate(make_clock(name, start, timedelta(days=1)))
    p.write_stream(
        clock,
        stream_path=f"clock-{name}",
        schema=SCHEMA,
        description="Clock that ticks every day",
    )

    p.main()
