import uuid

from beneath import __version__
from beneath.base import BaseClient
from beneath.stream import Stream
from beneath.proto import engine_pb2
from beneath.proto import gateway_pb2
from beneath.utils import datetime_to_ms
from beneath.utils import format_entity_name


class Client(BaseClient):
  """
  Client for interacting with the Beneath Data server.
  """

  def read_latest_batch(self, instance_id, limit, before=None):
    response = self.stub.ReadLatestRecords(
      gateway_pb2.ReadLatestRecordsRequest(
        instance_id=instance_id.bytes,
        limit=limit,
        before=datetime_to_ms(before) if before else 0,
      ), metadata=self.request_metadata
    )
    return response.records


  def read_batch(self, instance_id, where, limit, after):
    response = self.stub.ReadRecords(
      gateway_pb2.ReadRecordsRequest(
        instance_id=instance_id.bytes,
        where=where,
        limit=limit,
        after=after,
      ), metadata=self.request_metadata
    )
    return response.records


  def write_batch(self, instance_id, encoded_records):
    self.stub.WriteRecords(
      engine_pb2.WriteRecordsRequest(
        instance_id=instance_id.bytes,
        records=encoded_records
      ), metadata=self.request_metadata
    )


  def stream(self, project_name, stream_name):
    """
    Returns a Stream object identifying a Beneath stream

    Args:
      project (str): Name of the project that contains the stream.
      stream (str): Name of the stream.
    """

    details = self._query_control(
      variables={
        'name': format_entity_name(stream_name),
        'projectName': format_entity_name(project_name),
      },
      query="""
        query Stream($name: String!, $projectName: String!) {
          stream(
            name: $name, 
            projectName: $projectName,
          ) {
            streamID
            name
            schema
            avroSchema
            keyFields
            batch
            project {
              name
            }
            currentStreamInstanceID
          }
        }
      """
    )

    current_instance_id = details['stream']['currentStreamInstanceID']
    if current_instance_id is not None:
      current_instance_id = uuid.UUID(hex=current_instance_id)

    return Stream(
      client=self,
      stream_id=uuid.UUID(hex=details['stream']['streamID']),
      project_name=details['stream']['project']['name'],
      stream_name=details['stream']['name'],
      schema=details['stream']['schema'],
      key_fields=details['stream']['keyFields'],
      avro_schema=details['stream']['avroSchema'],
      batch=details['stream']['batch'],
      current_instance_id=current_instance_id,
    )
