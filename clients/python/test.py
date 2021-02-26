# pipeline.py

# This pipeline generates a stream `ticks` with a record for every minute
# since 1st Jan 2021.

# To test locally:
#     python ./pipeline.py test
# To prepare for running:
#     python ./pipeline.py stage USERNAME/PROJECT/ticker
# To run (after stage):
#     python ./pipeline.py run USERNAME/PROJECT/ticker
# To teardown:
#     python ./pipeline.py teardown USERNAME/PROJECT/ticker

import asyncio
import beneath
from datetime import datetime, timedelta, timezone

start = datetime(year=2021, month=1, day=1, tzinfo=timezone.utc)


async def ticker(p: beneath.Pipeline):
    last_tick = await p.checkpoints.get("last", default=start)
    while True:
        now = datetime.now(tz=last_tick.tzinfo)
        next_tick = last_tick + timedelta(minutes=1)
        if next_tick >= now:
            yield beneath.PIPELINE_IDLE
            wait = next_tick - now
            await asyncio.sleep(wait.total_seconds())
        yield {"time": next_tick}
        await p.checkpoints.set("last", next_tick)
        last_tick = next_tick


if __name__ == "__main__":

    p = beneath.Pipeline(parse_args=True)
    p.description = "Pipeline that emits a tick for every minute since 1st Jan 2021"

    ticks = p.generate(ticker)
    p.write_stream(
        ticks,
        stream_path="ticks",
        schema="""
            type Tick @schema {
                time: Timestamp! @key
            }
        """,
    )

    p.main()
