"""
  Model for computing metrics on daily Dai activity
"""

import apache_beam as beam
from apache_beam.options.pipeline_options import PipelineOptions
from beneath import Client
from beneath.beam import ReadFromBeneath, WriteToBeneath
from datetime import datetime

def run():
  client = Client()
  p = beam.Pipeline(options=PipelineOptions())
  
  daily = (p
    | ReadFromBeneath(client.stream("maker", "dai-transfers"))
    | beam.FlatMap(beam.FlatMap(expand_transfer))
    | beam.GroupByKey()
  )

  results = {}

  results["transfers"] = (daily | beam.CombinePerKey(lambda vs: sum(
    1 for v in vs if v["address"] and v["value"] < 0
  )))

  # TODO
  # results["users"] = # count(distinct u.address) as users
  # results["senders"] = # count(distinct if(u.value < 0, u.address, null)) as senders
  # results["receivers"] = # count(distinct if(u.value > 0, u.address, null)) as receivers

  results["minted"] = (daily | beam.CombinePerKey(lambda vs: sum(
    -1 * v["value"]
    for v in vs
    if v["address"] == None and v["value"] < 0
  )))

  results["burned"] = (daily | beam.CombinePerKey(lambda vs: sum(
    v["value"]
    for v in vs
    if v["address"] == None and v["value"] > 0
  )))

  results["turnover"] = (daily | beam.CombinePerKey(lambda vs: sum(
    v["value"]
    for v in vs
    if v["value"] > 0
  )))

  (results
   | beam.CoGroupByKey()
   | beam.Map(lambda v: { "day": v[0], **v[1] })
   | WriteToBeneath(client.stream("maker", "dai-daily-activity"))
  )

  # Run pipeline
  result = p.run()
  result.wait_until_finish()

def expand_transfer(t):
  day = datetime.fromisoformat(t["time"]).replace(hour=0, minute=0, second=0, microsecond=0)
  return [
    (day, {"address": t["from"], "value": -1 * int(t["value"])}),
    (day, {"address": t["to"], "value": int(t["value"])}),
  ]

if __name__ == '__main__':
  run()
