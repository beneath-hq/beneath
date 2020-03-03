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
            schema
            avroSchema
            streamIndexes {
              fields
              primary
              normalize
            }
            external
            batch
            manual
            retentionSeconds
            instancesCreatedCount
            instancesCommittedCount
            currentStreamInstanceID
          }
        }
      """
    )
    return result['streamByID']

  async def find_by_project_and_name(self, project_name, stream_name):
    result = await self.conn.query_control(
      variables={
        'name': format_entity_name(stream_name),
        'projectName': format_entity_name(project_name),
      },
      query="""
        query StreamByProjectAndName($name: String!, $projectName: String!) {
          streamByProjectAndName(
            name: $name, 
            projectName: $projectName,
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
            schema
            avroSchema
            streamIndexes {
              fields
              primary
              normalize
            }
            external
            batch
            manual
            retentionSeconds
            instancesCreatedCount
            instancesCommittedCount
            currentStreamInstanceID
          }
        }
      """
    )
    return result['streamByProjectAndName']

  async def create(self, schema, project_id, manual=False, batch=False):
    result = await self.conn.query_control(
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
            createdOn
            updatedOn
            project {
              projectID
              name
            }
            schema
            avroSchema
            streamIndexes {
              fields
              primary
              normalize
            }
            external
            batch
            manual
            retentionSeconds
            instancesCreatedCount
            instancesCommittedCount
            currentStreamInstanceID
          }
        }
      """
    )
    return result['createExternalStream']

  async def update(self, stream_id, schema=None, manual=None):
    variables = {'streamID': stream_id}
    if schema:
      variables['schema'] = schema
    if manual:
      variables['manual'] = bool(manual)

    result = await self.conn.query_control(
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
            createdOn
            updatedOn
            project {
              projectID
              name
            }
            schema
            avroSchema
            streamIndexes {
              fields
              primary
              normalize
            }
            external
            batch
            manual
            retentionSeconds
            instancesCreatedCount
            instancesCommittedCount
            currentStreamInstanceID
          }
        }
      """
    )
    return result['updateExternalStream']

  async def delete(self, stream_id):
    result = await self.conn.query_control(
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

  async def create_batch(self, stream_id):
    result = await self.conn.query_control(
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

  async def commit_batch(self, instance_id):
    result = await self.conn.query_control(
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

  async def clear_pending_batches(self, stream_id):
    result = await self.conn.query_control(
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
