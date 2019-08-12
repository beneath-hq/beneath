import base64
import apache_beam as beam
from apache_beam.io.gcp.internal.clients import bigquery
from apache_beam.options.pipeline_options import PipelineOptions
import sys
sys.path.append(
    '/Users/ericgreen/Desktop/Beneath/code/beneath-core/beneath-python')
from beneath.client import Client
from beneath.beam.beamio import WriteToBeneath

# model for extracting MKR transfers
query = '''
create temporary function hexToDecimal(x string)
returns string
language js as """
return BigInt(x).toString();
""";

with logs as (
  select l.transaction_hash as transaction_hash, l.block_timestamp as time, l.log_index as index_in_block, l.topics, l.data
  from `bigquery-public-data.crypto_ethereum.logs` l
  where l.address = '0x9f8f72aa9304c8b593d555f12ef6589cc3a579a2'  
  and l.block_timestamp > timestamp(date(2019, 7, 1))
)
select
  *
from (
  select
    l.transaction_hash,
    unix_millis(l.time) as _time,
    l.index_in_block as _index,
    'Transfer' as _name,
    concat('0x', substr(l.topics[offset(1)], 27)) as _from,
    concat('0x', substr(l.topics[offset(2)], 27)) as _to,
    hexToDecimal(l.data) as _value
  from logs l
  where l.topics[offset(0)] = '0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef' -- Transfer
) union all (
  select
    l.transaction_hash,
    unix_millis(l.time) as _time,
    l.index_in_block as _index,
    'Mint' as _name,
    null as _from,
    concat('0x', substr(l.topics[offset(1)], 27)) as _to,
    hexToDecimal(l.data) as _value
  from logs l
  where l.topics[offset(0)] = '0x0f6798a560793a54c3bcfe86a93cde1e73087d944c0ea20544137d4121396885' -- Mint
) union all (
  select
    l.transaction_hash,
    unix_millis(l.time) as _time,
    l.index_in_block as _index,
    'Burn' as _name,
    concat('0x', substr(l.topics[offset(1)], 27)) as _from,
    null as _to,
    hexToDecimal(l.data) as _value
  from logs l
  where l.topics[offset(0)] = '0xcc16f5dbb4873280815c1ee09dbd06736cffcc184412cf7a71a0fdb75d397ca5' -- Burn
)
'''

# # BigQuery returns all timestamps as strings, but our avro codec is expecting integers (# of milliseconds from epoch). This function reverts this interpretation for a specified column.
# def fix_timestamp(field):

# run the script
def run(argv=None):
  # set up beneath client
  # client = Client(secret='ZTmjxFZ2wBHG7izu+HBbjMZLLvfDD/L61r2MasM8kIg=')
  client = Client(secret='+qdMrH3BdNxACwX4FbldVsp3O2MrhrpMaXAce/8bems=')

  # get the stream where the data will be insert
  stream = client.get_stream("ethereum-3", "mkr-transfers")

  # create Beam pipeline
  options = PipelineOptions()
  pipeline = beam.Pipeline(options=options)
  (
      pipeline
      | 'Read' >> beam.io.Read(beam.io.BigQuerySource(query=query, use_standard_sql=True))
      | 'Write' >> WriteToBeneath(stream)
  )

  # Run the pipeline
  result = pipeline.run()
  result.wait_until_finish()

if __name__ == '__main__':
  run()
