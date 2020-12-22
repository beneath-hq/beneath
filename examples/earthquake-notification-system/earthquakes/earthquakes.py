"""
This example contains a pipeline that generates a stream, "earthquakes".

To run this example:
  1. Install the Beneath CLI:
    pip install beneath

  2. Generate a secret on beneath.dev and authenticate with the CLI:
    beneath auth SECRET

  3. Create a project for your data:
    beneath project create USERNAME/earthquakes
  
  4. Stage the pipeline
    python earthquakes.py stage USERNAME/earthquakes/get-earthquakes

  5. Run the pipeline
    python earthquakes.py run USERNAME/earthquakes/get-earthquakes

  6. Check out the data on beneath.dev!
"""

import asyncio
import aiohttp
import beneath
from datetime import datetime

BASE_URL = "https://earthquake.usgs.gov/fdsnws/event/1/query?format=geojson&starttime="
POLL_SECONDS = 30

# create a Beneath pipeline
p = beneath.Pipeline(parse_args=True)

# define a generator function that:
# -- continually queries the earthquake API
# -- "yields" data
# -- checkpoints its progress
async def query_earthquakes(p):
  # upon startup, the pipeline will fetch its most recent checkpoint from Beneath
  # upon FIRST startup, the pipeline will use the "default" value
  # see line 65 where we set the checkpoint
  checkpoint = await p.get_checkpoint("time", default=1603326090390)  # 22/10/20
  
  # create a session
  async with aiohttp.ClientSession() as session:
    # run forever
    while True:

      # construct the query URL from the checkpoint value
      starttime = datetime.fromtimestamp(checkpoint / 1000.0 + 1).isoformat()
      URL = BASE_URL + starttime
      
      # submit an asyncrhonous http request to the earthquake API
      async with session.get(URL) as resp:
        # get the json object from the http response
        data = await resp.json()

        # yield the earthquake data (it's nested within the "data" object)
        for earthquake in data['features']:
          yield earthquake['properties']

        # if we've received new data, set a new checkpoint, which is the time of the most recent earthquake
        if len(data['features']) > 0:
          checkpoint = data['features'][0]['properties']['time']
          await p.set_checkpoint("time", checkpoint)

        # sleep until next query
        await asyncio.sleep(POLL_SECONDS)

# add the generator as the first step in the pipeline
earthquakes = p.generate(query_earthquakes)

# write to Beneath as the second step in the pipeline
# define the stream's schema
p.write_stream(earthquakes, "earthquakes", """
  " Earthquakes fetched from https://earthquake.usgs.gov/. Code and how to deploy: https://gitlab.com/beneath-hq/beneath/-/tree/master/clients/python/examples/earthquake-notification-system/earthquakes"
  type Earthquake @stream @key(fields: "time") {
    " Time of earthquake "
    time: Timestamp!

    " Richter scale magnitude " 
    mag: Float32

    " Location of earthquake "
    place: String!

    " Link for more information "
    detail: String
  }
""")

# run the pipeline
p.main()
