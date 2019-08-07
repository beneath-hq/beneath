import io
import grpc
import uuid
import pandas as pd
import json
import time
import apache_beam as beam
from fastavro import schemaless_writer, schemaless_reader, reader, parse_schema
from beneath.stream import Stream
from beneath.proto import engine_pb2
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
      raise TypeError("secret must be a string")
    
    self._prepare()

  def __getstate__(self):
    return {
      "secret": self.secret,
    }  

  def __setstate__(self, obj):
    self.secret = obj["secret"]
    self._prepare()

  # create a client with the provided secret
  def _prepare(self):
    self.request_metadata = [('authorization', 'Bearer {}'.format(self.secret))]

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
    return Stream(
      client=self,
      project_name=details.project_name,
      stream_name=details.stream_name,
      current_instance_id=uuid.UUID(bytes=details.current_instance_id),
      avro_schema=parse_schema(json.loads(details.avro_schema)),
      batch=details.batch,
    )
