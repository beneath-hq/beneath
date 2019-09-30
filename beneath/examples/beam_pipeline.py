import base64
import apache_beam as beam
from apache_beam.io.gcp.internal.clients import bigquery
from apache_beam.options.pipeline_options import PipelineOptions
from beneath.client import Client
from beneath.beam import ReadFromBeneath, WriteToBeneath, ConvertFromJSONTypes

# transformation
class SelectColsFn(beam.DoFn):
  def process(self, element):
    record = {'number': element['number'], 'hash': element['hash']}
    return [record]

def run(argv=None):
  # set up beneath client
  client = Client(secret='IlYcyF9oaYRXCRyJw74AOVki2jNw12wdhkf/QGPCkgQ=')

  # get the stream of interest
  stream = client.stream("ethereum-eric", "type-testing-stream")

  # create Beam pipeline
  options = PipelineOptions()
  pipeline = beam.Pipeline(options=options)
  (
    pipeline
    | 'Read'         >> ReadFromBeneath(stream) # beam.io.Read(beam.io.BigQuerySource(query='SELECT * FROM `beneathcrypto.ethereum_eric.labeled_addresses`', use_standard_sql=True))
    | 'ConvertTypes' >> ConvertFromJSONTypes({"hash": "bytes", "difficulty": "int", "timestamp": "timestamp", "extraData": "bytes", "miner": "bytes"})
    # | 'print' >> beam.Map(print)
    | 'Write'        >> WriteToBeneath(stream)
  )

  # Run the pipeline
  result = pipeline.run()
  result.wait_until_finish()

if __name__ == '__main__':
  run()

# old pipeline steps
# | 'CleanupBytes' >> beam.Map(fix_bytes("hash"))
# | 'selectCols'   >> beam.ParDo(SelectColsFn())

# # if stream doesn't yet exist, create new table: (bake this into WriteToBeneath)
# beneath.create_stream(project="ethereum-2", stream="my-new-stream", auth="[key]")
