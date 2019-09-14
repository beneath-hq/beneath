"""
  Model for computing metrics on daily blocks
"""

import apache_beam as beam
from apache_beam.options.pipeline_options import PipelineOptions
from beneath import Client
from beneath.beam import ReadFromBeneath, WriteToBeneath
from datetime import datetime


def run():
  client = Client()
  p = beam.Pipeline(options=PipelineOptions())

  (
    p
    | ReadFromBeneath(client.stream("ethereum", "blocks"))
    | beam.Map(expand_block)
    | beam.GroupByKey()
    | beam.CombinePerKey(lambda vs: len(vs))
    | beam.Map(lambda v: {"day": v[0], "count": [1]})
    | WriteToBeneath(client.stream("ethereum", "blocks-daily"))
  )

  result = p.run()
  result.wait_until_finish()


def expand_block(b):
  day = datetime.fromisoformat(b["timestamp"]).replace(
      hour=0, minute=0, second=0, microsecond=0)
  return (day, b)


if __name__ == '__main__':
  run()
