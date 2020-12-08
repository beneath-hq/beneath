"""
This example contains a pipeline that listens to the "earthquakes" stream, and uses Twilio to send an SMS notification when big earthquakes happen.

To run this example:  
  1. Stage the pipeline
    python earthquake_notification.py stage demos/earthquakes/notify

  2. Run the pipeline
    python earthquake_notification.py run demos/earthquakes/notify
"""

import beneath
import os
import pytz
from datetime import datetime, timedelta
from twilio.rest import Client

TWILIO_ACCOUNT_ID = os.getenv("TWILIO_ACCOUNT_ID", default=None)
TWILIO_AUTH_TOKEN = os.getenv("TWILIO_AUTH_TOKEN", default=None)
TWILIO_FROM_PHONE_NUMBER = os.getenv("TWILIO_FROM_PHONE_NUMBER", default=None)
TO_PHONE_NUMBER = os.getenv("TO_PHONE_NUMBER", default=None)

CURRENT_EARTHQUAKE_THRESHOLD = 15 # minutes
BIG_EARTHQUAKE_THRESHOLD = 5.5 # on the Richter scale

# create a twilio client
twilio_client = Client(TWILIO_ACCOUNT_ID, TWILIO_AUTH_TOKEN)

# create a Beneath pipeline
p = beneath.Pipeline(parse_args=True)

async def detect_big_earthquakes(earthquake):
  # only react to big earthquakes
  is_big = earthquake['mag'] is not None and earthquake['mag'] > BIG_EARTHQUAKE_THRESHOLD
  # only react to current earthquakes
  is_current = datetime.now(tz=pytz.utc) - earthquake['time'] < timedelta(minutes=CURRENT_EARTHQUAKE_THRESHOLD)
  
  if is_big and is_current:
    twilio_client.messages.create(
      body=f"Earthquake Alert! \n \
            Time: {earthquake['time']} \n \
            Magnitude: {earthquake['mag']} \n \
            Location: {earthquake['place']}",
      from_=TWILIO_FROM_PHONE_NUMBER,
      to=TO_PHONE_NUMBER
    )
  
# read from Beneath as the first step in the Pipeline
earthquakes = p.read_stream("demos/earthquakes/earthquakes")

# add the notification logic as the second step in the pipeline
p.apply(earthquakes, detect_big_earthquakes)

# run the pipeline
p.main()

