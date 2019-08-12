import base64
import apache_beam as beam
from apache_beam.io.gcp.internal.clients import bigquery
from apache_beam.options.pipeline_options import PipelineOptions
import sys
sys.path.append(
    '/Users/ericgreen/Desktop/Beneath/code/beneath-core/beneath-python')
from beneath.client import Client
from beneath.beam.beamio import WriteToBeneath

# model for counting daily active addresses
query = '''
with
  senders as (
    select _time, _from as _address, -1 * cast(_value as numeric) as _value
    from `beneathcrypto.ethereum_3.dai_transfers_3_9da2ea8b`
  ),
  receivers as (
    select _time, _to as _address, cast(_value as numeric) as _value
    from `beneathcrypto.ethereum_3.dai_transfers_3_9da2ea8b`
  ),
  users as (
    select * from (select * from senders union all select * from receivers)
  )
select *
from (
  select
    unix_millis(timestamp_trunc(u._time, DAY)) as time,
    count(distinct u._address) as users,
    count(distinct if(u._value < 0, u._address, null)) as senders,
    count(distinct if(u._value > 0, u._address, null)) as receivers,
    countif(u._value < 0) as transfers,
    sum(if(u._address is null and u._value < 0, abs(u._value), 0)) / pow(10, 18) as minted,
    sum(if(u._address is null and u._value > 0, u._value, 0)) / pow(10, 18) as burned,
    sum(if(u._value > 0, u._value, 0)) / pow(10, 18) as turnover
  from users u
  group by time
  order by time
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
  stream = client.get_stream("ethereum-3", "dai-daily-active-addresses-3")

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
