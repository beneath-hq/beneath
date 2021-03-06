import asyncio
import aiohttp
from datetime import datetime, timedelta

BASE_URL = "https://earthquake.usgs.gov/fdsnws/event/1/query?format=geojson&orderby=time-asc"
POLL_SECONDS = 30

# This function does three things:
# 1. Continually queries the USGS earthquake API
# 2. Yields data
# 3. Checkpoints its progress
async def generate_earthquakes(p):
    # Upon startup, the pipeline will fetch its most recent checkpoint from Beneath
    # Upon *first* startup, the pipeline will use the "default" value
    # See line 39 where we set the checkpoint
    checkpoint = await p.checkpoints.get("time", default=1612137600000)  # 1/2/2021

    # Create a http session
    async with aiohttp.ClientSession() as session:
        # Run forever
        while True:

            # Construct the query URL from the checkpoint value
            # We fetch at most 7 days of data at a time (the API breaks if the result is +20k rows)
            starttime = datetime.fromtimestamp(checkpoint / 1000.0 + 1)
            endtime = starttime + timedelta(days=7)
            URL = f"{BASE_URL}&starttime={starttime.isoformat()}&endtime={endtime}"

            # Submit an asyncrhonous http request to the earthquake API
            async with session.get(URL) as resp:
                # Get the json object from the http response
                data = await resp.json()

                # Yield all the earthquake data (it's nested within the "data" object)
                for earthquake in data["features"]:
                    yield earthquake["properties"]

                # If we've received new data, set a new checkpoint, which is the time of the most recent earthquake
                if len(data["features"]) > 0:
                    checkpoint = data["features"][-1]["properties"]["time"]
                    await p.checkpoints.set("time", checkpoint)

                # Sleep until next query (unless we're catching up)
                if datetime.now() < endtime:
                    await asyncio.sleep(POLL_SECONDS)
