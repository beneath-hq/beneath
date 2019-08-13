"""
  Batch model for extracting Dai transfers into Beneath
"""

import apache_beam as beam
from apache_beam.options.pipeline_options import PipelineOptions
from beneath.client import Client
from beneath.beam import WriteToBeneath

with open("query.sql", "r") as f:
  query = f.read()

def run():
  # Connect to Beneath
  client = Client()

  # Define pipeline steps
  p = beam.Pipeline(options=PipelineOptions())
  (p
    | 'Read' >> beam.io.Read(beam.io.BigQuerySource(query=query, use_standard_sql=True))
    | 'Write' >> WriteToBeneath(client.stream("maker", "dai-transfers"))
  )

  # Run pipeline
  result = p.run()
  result.wait_until_finish()

if __name__ == '__main__':
  run()