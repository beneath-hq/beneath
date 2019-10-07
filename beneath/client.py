import os
import uuid
import warnings

from datetime import datetime
import grpc
import requests

from beneath import __version__
from beneath import config
from beneath.stream import Stream
from beneath.proto import engine_pb2
from beneath.proto import gateway_pb2
from beneath.proto import gateway_pb2_grpc
from beneath.utils import datetime_to_ms

class GraphQLError(Exception):
  def __init__(self, message, errors):
    super().__init__(message)
    self.errors = errors


class BeneathError(Exception):
  pass


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
      self.secret = os.getenv("BENEATH_SECRET", default=None)
    if self.secret is None:
      self.secret = config.read_secret()
    if not isinstance(self.secret, str):
      raise TypeError("secret must be a string")
    self.secret = self.secret.strip()

    self.channel = None
    self.request_metadata = None
    self.stub = None
    self._prepare()


  def __getstate__(self):
    return {"secret": self.secret}


  def __setstate__(self, obj):
    self.secret = obj["secret"]
    self._prepare()


  def _prepare(self):
    """ Called either in __init__ or after unpickling """
    self._connect_grpc()
    self._check_auth_and_version()


  def _connect_grpc(self):
    self.request_metadata = [('authorization', 'Bearer {}'.format(self.secret))]
    insecure = "localhost" in config.BENEATH_GATEWAY_HOST_GRPC
    if insecure:
      self.channel = grpc.insecure_channel(
        target=config.BENEATH_GATEWAY_HOST_GRPC,
        compression=grpc.Compression.Gzip,
      )
    else:
      self.channel = grpc.secure_channel(
        target=config.BENEATH_GATEWAY_HOST_GRPC,
        credentials=grpc.ssl_channel_credentials(),
        compression=grpc.Compression.Gzip,
      )
    self.stub = gateway_pb2_grpc.GatewayStub(self.channel)


  def _send_client_ping(self):
    return self.stub.SendClientPing(gateway_pb2.ClientPing(
      client_id=config.PYTHON_CLIENT_ID,
      client_version=__version__,
    ), metadata=self.request_metadata)


  def _check_auth_and_version(self):
    pong = self._send_client_ping()
    self._check_pong_status(pong)
    if not pong.authenticated:
      raise BeneathError("You must authenticate with 'beneath auth'")


  @classmethod
  def _check_pong_status(cls, pong):
    if pong.status == "warning":
      warnings.warn(
        "This version ({}) of the Beneath python library will soon be deprecated (recommended: {}). "
        "Update with 'pip install --upgrade beneath'.".format(__version__, pong.recommended_version)
      )
    elif pong.status == "deprecated":
      raise Exception(
        "This version ({}) of the Beneath python library is out-of-date (recommended: {}). "
        "Update with 'pip install --upgrade beneath' to continue.".format(__version__, pong.recommended_version)
      )


  @classmethod
  def _format_resource_name(cls, name):
    return name.replace("-", "_")


  @classmethod
  # time must be provided in datetime format in UTC
  def _format_time(cls, time):
    return time.isoformat() + 'Z'


  def _query_control(self, query, variables):
    """ Sends a GraphQL query to the control server """
    url = config.BENEATH_CONTROL_HOST + '/graphql'
    headers = {'Authorization': 'Bearer ' + self.secret}
    response = requests.post(url, headers=headers, json={
      'query': query,
      'variables': variables,
    })
    response.raise_for_status()
    obj = response.json()
    if 'errors' in obj:
      raise GraphQLError(obj['errors'][0]['message'], obj['errors'])
    return obj['data']


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
    details = self.get_stream_details(self._format_resource_name(project_name), self._format_resource_name(stream_name))
    return Stream(
      client=self,
      project_name=details['project']['name'],
      stream_name=details['name'],
      schema=details['schema'],
      key_fields=details['keyFields'],
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
    me = result['me']
    if me is None:
      raise Exception("Cannot call get_me when authenticated with a project key")
    return me


  def get_user_by_id(self, user_id):
    result = self._query_control(
      variables={
        'userID': user_id,
      },
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
    return result['user']


  def get_project_by_name(self, name):
    result = self._query_control(
      variables={
        'name': self._format_resource_name(name),
      },
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
    return result['projectByName']


  def get_organization_by_name(self, name):
    result = self._query_control(
      variables={
        'name': self._format_resource_name(name),
      },
      query="""
        query OrganizationByName($name: String!) {
          organizationByName(name: $name) {
            organizationID
            name
            createdOn
            updatedOn
            services {
              serviceID
              name
              kind
              readBytesQuota
              writeBytesQuota
            }
            users {
              userID
              username
              name
              createdOn
              readQuota
              writeQuota
            }
          }
        }
      """
    )
    return result['organizationByName']


  def get_model_details(self, project_name, model_name):
    result = self._query_control(
      variables={
        'name': model_name,
        'projectName': self._format_resource_name(project_name),
      },
      query="""
        query Model($name: String!, $projectName: String!) {
          model(name: $name, projectName: $projectName) {
            modelID
            name
            description
            sourceURL
            kind
            createdOn
            updatedOn
            inputStreams {
              streamID
              name
            }
            outputStreams {
              streamID
              name
            }
          }
        }
      """
    )
    return result['model']


  def get_usage(self, user_id, period=None, from_time=None, until=None):
    today = datetime.today()

    if (period is None) | (period == 'M'):
      default_time = datetime(today.year, today.month, 1)
    if period == 'H':
      default_time = datetime(today.year, today.month, today.day, today.hour)

    result = self._query_control(
      variables={
        'userID': user_id,
        'period': period if period else 'M',
        'from': self._format_time(from_time) if from_time else self._format_time(default_time),
        'until': self._format_time(until) if until else None
      },
      query="""
        query GetUserMetrics($userID: UUID!, $period: String!, $from: Time!, $until: Time) {
          getUserMetrics(userID: $userID, period: $period, from: $from, until: $until) {
            entityID
            period
            time
            readOps
            readBytes
            readRecords
            writeOps
            writeBytes
            writeRecords
          }
        }
      """
    )
    return result['getUserMetrics']


  def create_organization(self, name):
    result = self._query_control(
      variables={
        'name': self._format_resource_name(name),
      },
      query="""
        mutation CreateOrganization($name: String!) {
          createOrganization(name: $name) {
            organizationID
            name
            createdOn
            updatedOn
          }
        }
      """
    )
    return result['createOrganization']


  def create_project(self, name, display_name, organization_id, description=None, site_url=None, photo_url=None):
    result = self._query_control(
      variables={
        'name': self._format_resource_name(name),
        'displayName': display_name,
        'organizationID': organization_id,
        'description': description,
        'site': site_url,
        'photoURL': photo_url,
      },
      query="""
        mutation CreateProject($name: String!, $displayName: String, $organizationID: UUID!, $site: String, $description: String, $photoURL: String) {
          createProject(name: $name, displayName: $displayName, organizationID: $organizationID, site: $site, description: $description, photoURL: $photoURL) {
            projectID
            name
            displayName
            organizationID
            site
            description
            photoURL
            createdOn
            updatedOn
            users {
              userID
              name
              username
              photoURL
            }
          }
        }
      """
    )
    return result['createProject']


  def update_project(self, project_id, display_name, description=None, site_url=None, photo_url=None):
    result = self._query_control(
      variables={
        'projectID': project_id,
        'displayName': display_name,
        'description': description,
        'site': site_url,
        'photoURL': photo_url,
      },
      query="""
        mutation UpdateProject($projectID: UUID!, $displayName: String, $site: String, $description: String, $photoURL: String) {
          updateProject(projectID: $projectID, displayName: $displayName, site: $site, description: $description, photoURL: $photoURL) {
            projectID
            displayName
            site
            description
            photoURL
            updatedOn
          }
        }
      """
    )
    return result['updateProject']


  def delete_project(self, project_id):
    result = self._query_control(
      variables={
        'projectID': project_id,
      },
      query="""
        mutation DeleteProject($projectID: UUID!) {
          deleteProject(projectID: $projectID)
        }
      """
    )
    return result['deleteProject']


  def create_service(self, name, organization_id, read_bytes_quota, write_bytes_quota):
    result = self._query_control(
      variables={
        'name': self._format_resource_name(name),
        'organizationID': organization_id,
        'readBytesQuota': read_bytes_quota,
        'writeBytesQuota': write_bytes_quota,
      },
      query="""
        mutation CreateService($name: String!, $organizationID: UUID!, $readBytesQuota: Int!, $writeBytesQuota: Int!) {
          createService(name: $name, organizationID: $organizationID, readBytesQuota: $readBytesQuota, writeBytesQuota: $writeBytesQuota) {
            serviceID
            name
            kind
            readBytesQuota
            writeBytesQuota
          }
        }
      """
    )
    return result['createService']


  def update_service(self, service_id, name, organization_id, read_bytes_quota, write_bytes_quota):
    result = self._query_control(
      variables={
        'serviceID': service_id,
        'name': self._format_resource_name(name),
        'organizationID': organization_id,
        'readBytesQuota': read_bytes_quota,
        'writeBytesQuota': write_bytes_quota,
      },
      query="""
        mutation UpdateService($serviceID: UUID!, $name: String, $organizationID: UUID, $readBytesQuota: Int, $writeBytesQuota: Int) {
          updateService(serviceID: $serviceID, name: $name, organizationID: $organizationID, readBytesQuota: $readBytesQuota, writeBytesQuota: $writeBytesQuota) {
            serviceID
            name
            kind
            readBytesQuota
            writeBytesQuota
          }
        }
      """
    )
    return result['updateService']


  def delete_service(self, service_id):
    result = self._query_control(
      variables={
        'serviceID': service_id,
      },
      query="""
        mutation DeleteService($serviceID: UUID!) {
          deleteService(serviceID: $serviceID)
        }
      """
    )
    return result['deleteService']


  def update_service_permissions(self, service_id, stream_id, read, write):
    result = self._query_control(
      variables={
        'serviceID': service_id,
        'streamID': stream_id,
        'read': read,
        'write': write,
      },
      query="""
        mutation UpdateServicePermissions($serviceID: UUID!, $streamID: UUID!, $read: Boolean!, $write: Boolean!) {
          updateServicePermissions(serviceID: $serviceID, streamID: $streamID, read: $read, write: $write) {
            serviceID
            streamID
            read
            write
          }
        }
      """
    )
    return result['updateServicePermissions']


  def issue_service_secret(self, service_id, description):
    result = self._query_control(
      variables={
        'serviceID': service_id,
        'description': description,
      },
      query="""
        mutation IssueServiceSecret($serviceID: UUID!, $description: String!) {
          issueServiceSecret(serviceID: $serviceID, description: $description) {
            secretString
          }
        }
      """
    )
    return result['issueServiceSecret']


  def list_service_secrets(self, service_id):
    result = self._query_control(
      variables={
        'serviceID': service_id,
      },
      query="""
        query SecretsForService($serviceID: UUID!) {
          secretsForService(serviceID: $serviceID) {
            secretID
            description
            prefix
            createdOn
            updatedOn
          }
        }
      """
    )
    return result['secretsForService']


  def revoke_secret(self, secret_id):
    result = self._query_control(
      variables={
        'secretID': secret_id,
      },
      query="""
        mutation RevokeSecret($secretID: UUID!) {
          revokeSecret(secretID: $secretID)
        }
      """
    )
    return result['revokeSecret']


  def create_model(self, project_id, name, kind, source_url, description, input_stream_ids, output_stream_schemas):
    result = self._query_control(
      variables={
        'input': {
          'projectID': project_id,
          'name': self._format_resource_name(name),
          'kind': kind,
          'sourceURL': source_url,
          'description': description,
          'inputStreamIDs': input_stream_ids,
          'outputStreamSchemas': output_stream_schemas,
        },
      },
      query="""
        mutation CreateModel($input: CreateModelInput!) {
          createModel(input: $input) {
            modelID
            name
            description
            sourceURL
            kind
            createdOn
            updatedOn
            project {
              projectID
              name
            }
            inputStreams {
              streamID
              name
            }
            outputStreams {
              streamID
              name
            }
          }
        }
      """
    )
    return result['createModel']


  def update_model(self, model_id, source_url, description, input_stream_ids, output_stream_schemas):
    result = self._query_control(
      variables={
        'input': {
          'modelID': model_id,
          'sourceURL': source_url,
          'description': description,
          'inputStreamIDs': input_stream_ids,
          'outputStreamSchemas': output_stream_schemas,
        },
      },
      query="""
        mutation UpdateModel($input: UpdateModelInput!) {
          updateModel(input: $input) {
            modelID
            name
            description
            sourceURL
            kind
            createdOn
            updatedOn
            project {
              projectID
              name
            }
            inputStreams {
              streamID
              name
            }
            outputStreams {
              streamID
              name
            }
          }
        }
      """
    )
    return result['updateModel']


  def delete_model(self, model_id):
    result = self._query_control(
      variables={
        'modelID': model_id,
      },
      query="""
        mutation DeleteModel($modelID: UUID!) {
          deleteModel(modelID: $modelID) 
        }
      """
    )
    return result['deleteModel']


  def get_stream_details(self, project_name, stream_name):
    result = self._query_control(
      variables={
        'name': self._format_resource_name(stream_name),
        'projectName': self._format_resource_name(project_name),
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
    return result['stream']


  def create_external_stream(self, project_id, schema, manual=False, batch=False):
    result = self._query_control(
      variables={
        'projectID': project_id,
        'schema': schema,
        'batch': batch,
        'manual': bool(manual),
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
    return result['createExternalStream']


  def update_external_stream(self, stream_id, schema=None, manual=None):
    variables = {'streamID': stream_id}
    if schema:
      variables['schema'] = schema
    if manual:
      variables['manual'] = bool(manual)

    result = self._query_control(
      variables=variables,
      query="""
        mutation UpdateExternalStream($streamID: UUID!, $schema: String, $manual: Boolean) {
          updateExternalStream(
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
    return result['updateExternalStream']

  def delete_external_stream(self, stream_id):
    result = self._query_control(
      variables={
        'streamID': stream_id,
      },
      query="""
        mutation DeleteExternalStream($streamID: UUID!) {
          deleteExternalStream(streamID: $streamID)
        }
      """
    )
    return result['deleteExternalStream']

  def create_external_stream_batch(self, stream_id):
    result = self._query_control(
      variables={
        'streamID': stream_id,
      },
      query="""
        mutation CreateExternalStreamBatch($streamID: UUID!) {
          createExternalStreamBatch(streamID: $streamID) {
            instanceID
          }
        }
      """
    )
    return result['createExternalStreamBatch']

  def commit_external_stream_batch(self, instance_id):
    result = self._query_control(
      variables={
        'instanceID': instance_id,
      },
      query="""
        mutation CommitExternalStreamBatch($instanceID: UUID!) {
          commitExternalStreamBatch(instanceID: $instanceID)
        }
      """
    )
    return result['commitExternalStreamBatch']

  def clear_pending_external_stream_batches(self, stream_id):
    result = self._query_control(
      variables={
        'streamID': stream_id,
      },
      query="""
        mutation ClearPendingExternalStreamBatches($streamID: UUID!) {
          clearPendingExternalStreamBatches(streamID: $streamID) 
        }
      """
    )
    return result['clearPendingExternalStreamBatches']

  def create_model_batch(self, model_id):
    result = self._query_control(
      variables={
        'modelID': model_id,
      },
      query="""
        mutation CreateModelBatch($modelID: UUID!) {
          createModelBatch(modelID: $modelID) {
            instanceID
            stream {
              streamID
            }
          }
        }
      """
    )
    return result['createModelBatch']

  def commit_model_batch(self, model_id, instance_ids):
    result = self._query_control(
      variables={
        'modelID': model_id,
        'instanceIDs': instance_ids,
      },
      query="""
        mutation CommitModelBatch($modelID: UUID!, $instanceIDs: [UUID!]!) {
          commitModelBatch(modelID: $modelID, instanceIDs: $instanceIDs)
        }
      """
    )
    return result['commitModelBatch']

  def clear_pending_model_batches(self, model_id):
    result = self._query_control(
      variables={
        'modelID': model_id,
      },
      query="""
        mutation ClearPendingModelBatches($modelID: UUID!) {
          clearPendingModelBatches(modelID: $modelID)
        }
      """
    )
    return result['clearPendingModelBatches']
