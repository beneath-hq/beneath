import base64
import apache_beam as beam
from apache_beam.io.gcp.internal.clients import bigquery
from apache_beam.options.pipeline_options import PipelineOptions
from beneath.client import Client
from beneath.beam.beamio import WriteToBeneath

# transformation
class SelectColsFn(beam.DoFn):
  def process(self, element):
    record = {'number': element['number'], 'hash': element['hash']}
    return [record]

# BigQuery interprets all bytes as base64-encoded. This function reverts this interpretation for a specified column.
def fix_bytes(field):
  def _fix_bytes(row):
    row[field] = base64.b64decode(row[field])
    return row
  return _fix_bytes

# BigQuery returns timestamps as strings, and fastavro expects them as integers (milliseconds from epoch). This function reverts this interpretation for a specified column.
# def fix_timestamps()

# run the script
def run(argv=None):
  # set up beneath client
  client = Client(secret='ZTmjxFZ2wBHG7izu+HBbjMZLLvfDD/L61r2MasM8kIg=')

  # get the stream of interest
  stream = client.get_stream("ethereum-2", "block-number-basics")

  # create Beam pipeline
  options = PipelineOptions()
  pipeline = beam.Pipeline(options=options)
  (
    pipeline
    | 'Read'         >> beam.io.Read(beam.io.BigQuerySource(query='SELECT * FROM `beneathcrypto.ethereum_2.block_number`', use_standard_sql=True))
    | 'CleanupBytes' >> beam.Map(fix_bytes("hash"))
    | 'selectCols'   >> beam.ParDo(SelectColsFn())
    # | 'transform'    >> beam.Map(print)
    | 'Write'        >> WriteToBeneath(stream)
  )

  # Run the pipeline
  result = pipeline.run()
  result.wait_until_finish()

if __name__ == '__main__':
  run()


########################################################

# from beneath.beam import ReadFromBeneath, WriteToBeneath

# client = Client(secret='ZTmjxFZ2wBHG7izu+HBbjMZLLvfDD/L61r2MasM8kIg=')

# stream_definition = """
#   type BlockNumberBasics @stream(name: "basics", key: "number") {
#     number: Int!
#     hash: Bytes32!
#   }
# """

#   (
#       pipeline
#       | 'Read' >> ReadFromBeneath(stream=client.get_stream("ethereum-2", "block-number"))
#       | ... user defined transform / group / ml model
#       | 'Write' >> WriteToBeneath(stream=client.get_stream("ethereum-2", "block-number-basics"))
#   )

# # if stream doesn't yet exist, create new table: (bake this into WriteToBeneath)
# beneath.create_stream(project="ethereum-2", stream="my-new-stream", auth="[key]")
