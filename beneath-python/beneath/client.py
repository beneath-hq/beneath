from fastavro import schemaless_writer, schemaless_reader, reader, parse_schema
import io
import grpc
import pandas as pd
from beneath.proto import engine_pb2
import sys
import json
import time

# if you don't insert the proto directory into your path, then the _pb files cannot import one another
sys.path.insert(
    1, '/Users/ericgreen/Desktop/Beneath/code/beneath-core/beneath-python/beneath/proto')

from beneath.proto import gateway_pb2
from beneath.proto import gateway_pb2_grpc 

# create a map on the client. instanceid -> schema, so we cache it to remember it
# create fn "getAvroSchema" to see if the schema is in memory already

class Client:
  """Client to bundle configuration for API requests.

  Args:
    secret (str):
      The user's password to authenticate permission to access Beneath. 
  """

  # initialize the client with the user's secret
  def __init__(self, secret):
    self.secret = secret
    if not isinstance(secret, str):
      raise Exception("secret must be a string")

    self.request_metadata = [('authorization', 'Bearer {}'.format(secret))]

    # open a grpc channel from the client to the server
    # TODO: create a SSL/TLS connection
    self.channel = grpc.insecure_channel('localhost:50051')

    # create a "stub" (aka a client). the stub has all the methods that the gateway server has. so it'll have ReadRecords(), WriteRecords(), and GetStreamDetails()
    self.stub = gateway_pb2_grpc.GatewayStub(self.channel)

    # create a dictionary to remember schemas
    self.avro_schemas = dict()

  # get a stream's details
  def get_stream(self, project_name, stream_name):
    details = self.stub.GetStreamDetails(
        gateway_pb2.StreamDetailsRequest(
            project_name=project_name, stream_name=stream_name),
        metadata=self.request_metadata
    )

    # store the stream's schema in memory
    self.avro_schemas[details.current_instance_id] = details.avro_schema

    # return a Stream class
    return Stream(self, details)


class Stream:
  """Stream enables read/write operations.

  Args:
    client (Client):
      Authenticator to data on Beneath.
    details (StreamDetailsResponse):
      Contains all metadata related to a Stream.
  """

  def __init__(self, client, details):
    self.client = client
    self.project_name = details.project_name
    self.stream_name = details.stream_name
    self.current_instance_id = details.current_instance_id
    self.avro_schema = parse_schema(json.loads(details.avro_schema))

  # set default instance_id to current_instance_id, unless user specifies a different one
  def read_records(self, where, limit, instance_id=None):
    # unless specified otherwise, instance_id is the current_instance_id 
    if instance_id is None:
      instance_id = self.current_instance_id

    # gRPC ReadRecords from gateway
    response = self.client.stub.ReadRecords(
        gateway_pb2.ReadRecordsRequest(instance_id=instance_id, where=self._parse_where(where), limit=limit), metadata=self.client.request_metadata)

    # decode avro
    # 1. is there a way to parallelize this? response.records is an irregular object
    # 2. should I not be closing the reader every time?
    decoded_data = [0]*len(response.records)
    for i in range(len(response.records)):
      decoded_data[i] = self._decode_avro(response.records[i].avro_data)

    # return pandas dataframe
    df = pd.DataFrame(decoded_data)
    return df

  def write_record(self, instance_id, record, sequence_number=None):
    # ensure record is a dict
    if not isinstance(record, dict):
      raise Exception("record must be a dict")

    # encode avro
    encoded_data = self._encode_avro(record)
    if sequence_number is None:
      sequence_number = int(round(time.time() * 1000))
    new_record = engine_pb2.Record(avro_data=encoded_data, sequence_number=sequence_number)

    # gRPC WriteRecords to gateway
    response = self.client.stub.WriteRecords(
      engine_pb2.WriteRecordsRequest(instance_id=instance_id, records=[new_record]), metadata=self.client.request_metadata)
    return response
    
  def _decode_avro(self, data):
    reader = io.BytesIO(data)
    record = schemaless_reader(reader, self.avro_schema)
    reader.close()
    return record

  def _encode_avro(self, record):
    writer = io.BytesIO()
    schemaless_writer(writer, self.avro_schema, record)
    result = writer.getvalue()
    writer.close()
    return result

  def _parse_where(self, where):
    if isinstance(where, str):
      return where
    elif isinstance(where, dict):
      return json.dumps(where)
    else:
      raise Exception("expected json string or dict for parameter 'where'")
