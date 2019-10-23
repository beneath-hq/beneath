from datetime import datetime

from beneath import __version__
from beneath.base import BaseClient
from beneath.utils import format_entity_name
from beneath.utils import format_graphql_time


class AdminClient(BaseClient):
  """
  Client for interacting with the Beneath Admin server.
  """

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


  def get_user_by_username(self, username):
    result = self._query_control(
      variables={
        'username': username,
      },
      query="""
        query UserByUsername($username: String!) {
          userByUsername(
            username: $username
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
    return result['userByUsername']


  def get_project_by_name(self, name):
    result = self._query_control(
      variables={
        'name': format_entity_name(name),
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
        'name': format_entity_name(name),
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
              readQuota
              writeQuota
            }
            users {
              userID
              username
              name
              createdOn
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
        'projectName': format_entity_name(project_name),
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
    if (period is None) or (period == 'M'):
      default_time = datetime(today.year, today.month, 1)
    elif period == 'H':
      default_time = datetime(today.year, today.month, today.day, today.hour)

    result = self._query_control(
      variables={
        'userID': user_id,
        'period': period if period else 'M',
        'from': format_graphql_time(from_time) if from_time else format_graphql_time(default_time),
        'until': format_graphql_time(until) if until else None
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
        'name': format_entity_name(name),
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


  def add_user_to_organization(self, username, organization_id, view, admin):
    result = self._query_control(
      variables={
        'username': username,
        'organizationID': organization_id,
        'view': view,
        'admin': admin,
      },
      query="""
        mutation AddUserToOrganization($username: String!, $organizationID: UUID!, $view: Boolean!, $admin: Boolean!) {
          addUserToOrganization(username: $username, organizationID: $organizationID, view: $view, admin: $admin) {
            userID
            username
            name
            createdOn
            projects {
              name
            }
            readQuota
            writeQuota
          }
        }
      """
    )
    return result['addUserToOrganization']


  def rm_user_from_organization(self, user_id, organization_id):
    result = self._query_control(
      variables={
        'userID': user_id,
        'organizationID': organization_id,
      },
      query="""
        mutation RemoveUserFromOrganization($userID: UUID!, $organizationID: UUID!) {
          removeUserFromOrganization(userID: $userID, organizationID: $organizationID)
        }
      """
    )
    return result['removeUserFromOrganization']


  def create_project(self, name, display_name, organization_id, description=None, site_url=None, photo_url=None):
    result = self._query_control(
      variables={
        'name': format_entity_name(name),
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


  def add_user_to_project(self, username, project_id, view, create, admin):
    result = self._query_control(
      variables={
        'username': username,
        'projectID': project_id,
        'view': view,
        'create': create,
        'admin': admin,
      },
      query="""
        mutation AddUserToProject($username: String!, $projectID: UUID!, $view: Boolean!, $create: Boolean!, $admin: Boolean!) {
          addUserToProject(username: $username, projectID: $projectID, view: $view, create: $create, admin: $admin) {
            userID
            username
            name
            createdOn
            projects {
              name
            }
            readQuota
            writeQuota
          }
        }
      """
    )
    return result['addUserToProject']


  def rm_user_from_project(self, user_id, project_id):
    result = self._query_control(
      variables={
        'userID': user_id,
        'projectID': project_id,
      },
      query="""
        mutation RemoveUserFromProject($userID: UUID!, $projectID: UUID!) {
          removeUserFromProject(userID: $userID, projectID: $projectID)
        }
      """
    )
    return result['removeUserFromProject']


  def get_service_by_name_and_organization(self, name, organization_name):
    result = self._query_control(
      variables={
        'name': format_entity_name(name),
        'organizationName': format_entity_name(organization_name),
      },
      query="""
        query ServiceByNameAndOrganization($name: String!, $organizationName: String!) {
          serviceByNameAndOrganization(name: $name, organizationName: $organizationName) {
            serviceID
            name
            kind
            readQuota
            writeQuota
          }
        }
      """
    )
    return result['serviceByNameAndOrganization']


  def create_service(self, name, organization_id, read_quota_bytes, write_quota_bytes):
    result = self._query_control(
      variables={
        'name': format_entity_name(name),
        'organizationID': organization_id,
        'readQuota': read_quota_bytes,
        'writeQuota': write_quota_bytes,
      },
      query="""
        mutation CreateService($name: String!, $organizationID: UUID!, $readQuota: Int!, $writeQuota: Int!) {
          createService(name: $name, organizationID: $organizationID, readQuota: $readQuota, writeQuota: $writeQuota) {
            serviceID
            name
            kind
            readQuota
            writeQuota
          }
        }
      """
    )
    return result['createService']


  def update_service(self, service_id, name, read_quota_bytes, write_quota_bytes):
    result = self._query_control(
      variables={
        'serviceID': service_id,
        'name': format_entity_name(name) if name is not None else None,
        'readQuota': read_quota_bytes,
        'writeQuota': write_quota_bytes,
      },
      query="""
        mutation UpdateService($serviceID: UUID!, $name: String, $readQuota: Int, $writeQuota: Int) {
          updateService(serviceID: $serviceID, name: $name, readQuota: $readQuota, writeQuota: $writeQuota) {
            serviceID
            name
            kind
            readQuota
            writeQuota
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
        mutation UpdateServicePermissions($serviceID: UUID!, $streamID: UUID!, $read: Boolean, $write: Boolean) {
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
          'name': format_entity_name(name),
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
    