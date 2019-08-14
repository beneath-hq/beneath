import grpc
import uuid
import requests
import warnings

from beneath import __version__
from beneath import config
from beneath.stream import Stream
from beneath.proto import engine_pb2
from beneath.proto import gateway_pb2
from beneath.proto import gateway_pb2_grpc

class Client:
  """
  Client to bundle configuration for API requests.
  """

  def __init__(self, secret=None):
    """
    Args:
      secret (str): A beneath secret to use for authentication. If not set, reads secret from ~/.beneath.
    """
    self.secret = secret
    if self.secret is None:
      self.secret = config.read_secret()
    if not isinstance(self.secret, str):
      raise TypeError("secret must be a string")    
    self._prepare()


  def __getstate__(self):
    return { "secret": self.secret }


  def __setstate__(self, obj):
    self.secret = obj["secret"]
    self._prepare()


  def _prepare(self):
    """ Called either in __init__ or after unpickling """
    self._connect_grpc()
    self._check_auth_and_version()


  def _connect_grpc(self):
    self.request_metadata = [('authorization', 'Bearer {}'.format(self.secret))]
    self.channel = grpc.insecure_channel(config.BENEATH_GATEWAY_HOST_GRPC)
    self.stub = gateway_pb2_grpc.GatewayStub(self.channel)


  def _send_client_ping(self):
    return self.stub.SendClientPing(gateway_pb2.ClientPing(
      client_id=config.PYTHON_CLIENT_ID,
      client_version=__version__,
    ), metadata=self.request_metadata)


  def _check_auth_and_version(self):
    pong = self._send_client_ping()
    self._check_pong_status(pong.status)
    if not pong.authenticated:
      raise Exception("You must authenticate with 'beneath auth' or by passing the 'secret' arg")


  def _check_pong_status(self, status):
    if status == "warning":
      warnings.warn(
        "This version of the Beneath python library will soon be deprecated."
        "Update with 'pip install --upgrade beneath'."
      )
    elif status == "deprecated":
      raise Exception(
        "This version of the Beneath python library is out-of-date."
        "Update with 'pip install --upgrade beneath' to continue."
      )


  def _query_control(self, query, variables):
    """ Sends a GraphQL query to the control server """
    headers = {"Authorization": "Bearer " + self.secret}
    response = requests.post(
      config.BENEATH_CONTROL_HOST + '/graphql', 
      json={'query': query, 'variables': variables},
      headers=headers
    )
    response.raise_for_status()
    return response.json()

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


  def stream(self, project_name, stream_name):
    """
    Returns a Stream object identifying a Beneath stream

    Args:
      project (str): Name of the project that contains the stream.
      stream (str): Name of the stream.
    """
    details = self.get_stream_details(project_name, stream_name)
    return Stream(
      client=self,
      project_name=details['project']['name'],
      stream_name=details['name'],
      schema=details['schema'],
      avro_schema=details['avroSchema'],
      batch=details['batch'],
      current_instance_id=uuid.UUID(hex=details['currentStreamInstanceID']),
    )


  def get_me(self):
    """
      Returns info about the authenticated user.
      Returns None if authenicated with a project secret.
    """
    result = self._query_control(
      variables={},
      query="""
        query Me {
          me {
            userID
            user {
              username
              name
            }
            email
            updatedOn
          }
        }
      """
    )
    me = result['data']['me']
    if me is None:
      raise Exception("Cannot call get_me when authenticated with a project key")
    return me


  def get_user_by_id(self, user_id):
    result = self._query_control(
      variables={ "userID": user_id },
      query="""
        query User($userID: UUID!) {
          user(
            userID: $userID
          ) {
            userID
            username
            name
            bio
            photoURL
            createdOn
            projects {
              name
              createdOn
              updatedOn
              streams {
                name
              }
            }
          }
        }
      """
    )
    return result['data']['user']


  def get_project_by_name(self, name):
    result = self._query_control(
        variables={ "name": name },
        query="""
          query ProjectByName($name: String!) {
            projectByName(name: $name) {
              projectID
              name
              displayName
              site
              description
              photoURL
              createdOn
              updatedOn
              users {
                username
              }
              streams {
                name
              }
            }
        }
      """
    )
    return result['data']['projectByName']


  def get_stream_details(self, project_name, stream_name):
    result = self._query_control(
      variables={
        'name': stream_name,
        'projectName': project_name,
      },
      query="""
        query Stream($name: String!, $projectName: String!) {
          stream(
            name: $name, 
            projectName: $projectName,
          ) {
            streamID
            name
            description
            schema
            avroSchema
            keyFields
            external
            batch
            manual
            project {
              name
            }
            currentStreamInstanceID
            createdOn
            updatedOn
          }
        }
      """
    )
    return result['data']['stream']
    

  def create_external_stream(self, project_id, schema, manual=None):
    result = self._query_control(
      variables={
        "projectID": project_id,
        "schema": schema,
        "batch": False,
        "manual": bool(manual),
      },
      query="""
        mutation CreateExternalStream($projectID: UUID!, $schema: String!, $batch: Boolean!, $manual: Boolean!) {
          createExternalStream(
            projectID: $projectID,
            schema: $schema,
            batch: $batch,
            manual: $manual
          ) {
            streamID
            name
            description
            schema
            avroSchema
            keyFields
            external
            batch
            manual
            project {
              projectID
              name
            }
            currentStreamInstanceID
            createdOn
            updatedOn
          }
        }
      """
    )
    return result


  def update_external_stream(self, stream_id, schema=None, manual=None):
    variables = { "streamID": stream_id }
    if schema != None:
      variables["schema"] = schema
    if manual != None:
      variables["manual"] = bool(manual)
      
    result = self._query_control(
      variables=variables,
      query="""
        mutation UpdateStream($streamID: UUID!, $schema: String, $manual: Boolean) {
          updateStream(
            streamID: $streamID,
            schema: $schema,
            manual: $manual
          ) {
            streamID
            name
            description
            schema
            avroSchema
            keyFields
            external
            batch
            manual
            project {
              projectID
              name
            }
            currentStreamInstanceID
            createdOn
            updatedOn
          }
        }
      """
    )
    return result
