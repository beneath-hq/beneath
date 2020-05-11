from beneath.connection import Connection
from beneath.utils import format_entity_name


class Streams:

  def __init__(self, conn: Connection):
    self.conn = conn

  async def find_by_id(self, stream_id):
    result = await self.conn.query_control(
      variables={
        'streamID': stream_id,
      },
      query="""
        query StreamByID($streamID: UUID!) {
          streamByID(streamID: $streamID) {
            streamID
            name
            description
            createdOn
            updatedOn
            project {
              projectID
              name
            }
            schemaKind
            schema
            avroSchema
            streamIndexes {
              fields
              primary
              normalize
            }
            retentionSeconds
            enableManualWrites
            primaryStreamInstanceID
            instancesCreatedCount
            instancesDeletedCount
            instancesMadeFinalCount
            instancesMadePrimaryCount
          }
        }
      """
    )
    return result['streamByID']

  async def find_by_organization_project_and_name(self, organization_name, project_name, stream_name):
    result = await self.conn.query_control(
      variables={
        'organizationName': format_entity_name(organization_name),
        'projectName': format_entity_name(project_name),
        'streamName': format_entity_name(stream_name),
      },
      query="""
        query StreamByOrganizationProjectAndName($organizationName: String!, $projectName: String!, $streamName: String!) {
          streamByOrganizationProjectAndName(
            organizationName: $organizationName,
            projectName: $projectName,
            streamName: $streamName, 
          ) {
            streamID
            name
            description
            createdOn
            updatedOn
            project {
              projectID
              name
            }
            schemaKind
            schema
            avroSchema
            streamIndexes {
              fields
              primary
              normalize
            }
            retentionSeconds
            enableManualWrites
            primaryStreamInstanceID
            instancesCreatedCount
            instancesDeletedCount
            instancesMadeFinalCount
            instancesMadePrimaryCount
          }
        }
      """
    )
    return result['streamByOrganizationProjectAndName']

  async def find_instances(self, stream_id):
    result = await self.conn.query_control(
      variables={
        'streamID': stream_id,
      },
      query="""
        query StreamInstancesForStream($streamID: UUID!) {
          streamInstancesForStream(streamID: $streamID) {
            streamInstanceID
            streamID
            createdOn
            madePrimaryOn
            madeFinalOn
          }
        }
      """
    )
    return result['streamInstancesForStream']

  async def stage(
    self,
    organization_name,
    project_name,
    stream_name,
    schema_kind,
    schema,
    retention_seconds=None,
    enable_manual_writes=None,
    create_primary_instance=None,
  ):
    result = await self.conn.query_control(
      variables={
        'organizationName': organization_name,
        'projectName': project_name,
        'streamName': stream_name,
        'schemaKind': schema_kind,
        'schema': schema,
        'retentionSeconds': retention_seconds,
        'enableManualWrites': enable_manual_writes,
        'createPrimaryStreamInstance': create_primary_instance,
      },
      query="""
        mutation stageStream(
          $organizationName: String!,
          $projectName: String!,
          $streamName: String!
          $schemaKind: StreamSchemaKind!,
          $schema: String!,
          $retentionSeconds: Int,
          $enableManualWrites: Boolean,
          $createPrimaryStreamInstance: Boolean,
        ) {
          stageStream(
            organizationName: $organizationName,
            projectName: $projectName,
            streamName: $streamName,
            schemaKind: $schemaKind,
            schema: $schema,
            retentionSeconds: $retentionSeconds,
            enableManualWrites: $enableManualWrites,
            createPrimaryStreamInstance: $createPrimaryStreamInstance,
          ) {
            streamID
            name
            description
            createdOn
            updatedOn
            project {
              projectID
              name
            }
            schemaKind
            schema
            avroSchema
            streamIndexes {
              fields
              primary
              normalize
            }
            retentionSeconds
            enableManualWrites
            primaryStreamInstanceID
            instancesCreatedCount
            instancesDeletedCount
            instancesMadeFinalCount
            instancesMadePrimaryCount
          }
        }
      """
    )
    return result['stageStream']

  async def delete(self, stream_id):
    result = await self.conn.query_control(
      variables={
        'streamID': stream_id,
      },
      query="""
        mutation DeleteStream($streamID: UUID!) {
          deleteStream(streamID: $streamID)
        }
      """
    )
    return result['deleteStream']

  async def create_instance(self, stream_id):
    result = await self.conn.query_control(
      variables={
        'streamID': stream_id,
      },
      query="""
        mutation createStreamInstance($streamID: UUID!) {
          createStreamInstance(streamID: $streamID) {
            streamInstanceID
            streamID
            createdOn
            madePrimaryOn
            madeFinalOn
          }
        }
      """,
    )
    return result['createStreamInstance']

  async def update_instance(self, instance_id, make_final=None, make_primary=None, delete_previous_primary=None):
    result = await self.conn.query_control(
      variables={
        'instanceID': instance_id,
        'makeFinal': make_final,
        'makePrimary': make_primary,
        'deletePreviousPrimary': delete_previous_primary,
      },
      query="""
        mutation updateStreamInstance($instanceID: UUID!, $makeFinal: Boolean, $makePrimary: Boolean, $deletePreviousPrimary: Boolean) {
          updateStreamInstance(
            instanceID: $instanceID,
            makeFinal: $makeFinal,
            makePrimary: $makePrimary,
            deletePreviousPrimary: $deletePreviousPrimary,
          ) {
            streamInstanceID
            streamID
            createdOn
            madePrimaryOn
            madeFinalOn
          }
        }
      """,
    )
    return result['updateStreamInstance']

  async def delete_instance(self, instance_id):
    result = await self.conn.query_control(
      variables={
        'instanceID': instance_id,
      },
      query="""
        mutation deleteStreamInstance($instanceID: UUID!) {
          deleteStreamInstance(instanceID: $instanceID)
        }
      """,
    )
    return result['deleteStreamInstance']
