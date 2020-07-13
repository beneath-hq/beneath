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
            allowManualWrites
            useLog
            useIndex
            useWarehouse
            logRetentionSeconds
            indexRetentionSeconds
            warehouseRetentionSeconds
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
            allowManualWrites
            useLog
            useIndex
            useWarehouse
            logRetentionSeconds
            indexRetentionSeconds
            warehouseRetentionSeconds
            primaryStreamInstance {
              streamInstanceID
              createdOn
              version
              madePrimaryOn
              madeFinalOn
            }
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
            version
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
    description=None,
    allow_manual_writes=None,
    use_log=None,
    use_index=None,
    use_warehouse=None,
    log_retention_seconds=None,
    index_retention_seconds=None,
    warehouse_retention_seconds=None,
  ):
    result = await self.conn.query_control(
      variables={
        'organizationName': format_entity_name(organization_name),
        'projectName': format_entity_name(project_name),
        'streamName': format_entity_name(stream_name),
        'schemaKind': schema_kind,
        'schema': schema,
        'description': description,
        'allowManualWrites': allow_manual_writes,
        'useLog': use_log,
        'useIndex': use_index,
        'useWarehouse': use_warehouse,
        'logRetentionSeconds': log_retention_seconds,
        'indexRetentionSeconds': index_retention_seconds,
        'warehouseRetentionSeconds': warehouse_retention_seconds,
      },
      query="""
        mutation stageStream(
          $organizationName: String!,
          $projectName: String!,
          $streamName: String!
          $schemaKind: StreamSchemaKind!,
          $schema: String!,
          $description: String,
          $allowManualWrites: Boolean,
          $useLog: Boolean,
          $useIndex: Boolean,
          $useWarehouse: Boolean,
          $logRetentionSeconds: Int,
          $indexRetentionSeconds: Int,
          $warehouseRetentionSeconds: Int,
        ) {
          stageStream(
            organizationName: $organizationName,
            projectName: $projectName,
            streamName: $streamName,
            schemaKind: $schemaKind,
            schema: $schema,
            description: $description,
            allowManualWrites: $allowManualWrites,
            useLog: $useLog,
            useIndex: $useIndex,
            useWarehouse: $useWarehouse,
            logRetentionSeconds: $logRetentionSeconds,
            indexRetentionSeconds: $indexRetentionSeconds,
            warehouseRetentionSeconds: $warehouseRetentionSeconds,
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
            allowManualWrites
            useLog
            useIndex
            useWarehouse
            logRetentionSeconds
            indexRetentionSeconds
            warehouseRetentionSeconds
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

  async def stage_instance(self, stream_id, version: int, make_final=None, make_primary=None):
    result = await self.conn.query_control(
      variables={
        "streamID": stream_id,
        "version": version,
        "makeFinal": make_final,
        "makePrimary": make_primary,
      },
      query="""
        mutation stageStreamInstance($streamID: UUID!, $version: Int!, $makeFinal: Boolean, $makePrimary: Boolean) {
          stageStreamInstance(
            streamID: $streamID,
            version: $version,
            makeFinal: $makeFinal,
            makePrimary: $makePrimary,
          ) {
            streamInstanceID
            streamID
            createdOn
            version
            madePrimaryOn
            madeFinalOn
          }
        }
      """,
    )
    return result['stageStreamInstance']

  async def update_instance(self, instance_id, make_final=None, make_primary=None):
    result = await self.conn.query_control(
      variables={
        'instanceID': instance_id,
        'makeFinal': make_final,
        'makePrimary': make_primary,
      },
      query="""
        mutation updateStreamInstance($instanceID: UUID!, $makeFinal: Boolean, $makePrimary: Boolean) {
          updateStreamInstance(
            instanceID: $instanceID,
            makeFinal: $makeFinal,
            makePrimary: $makePrimary,
          ) {
            streamInstanceID
            streamID
            createdOn
            version
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
