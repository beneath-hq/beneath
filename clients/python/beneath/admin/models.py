from beneath.connection import Connection
from beneath.utils import format_entity_name


class Models:

  def __init__(self, conn: Connection):
    self.conn = conn

  async def find_by_project_and_name(self, project_name, model_name):
    result = await self.conn.query_control(
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

  async def create(
    self,
    name,
    project_id,
    kind,
    source_url,
    description,
    input_stream_ids,
    output_stream_schemas,
  ):
    result = await self.conn.query_control(
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

  async def update(
    self,
    model_id,
    source_url,
    description,
    input_stream_ids,
    output_stream_schemas,
  ):
    result = await self.conn.query_control(
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

  async def delete(self, model_id):
    result = await self.conn.query_control(
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

  async def create_batch(self, model_id):
    result = await self.conn.query_control(
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

  async def commit_batch(self, model_id, instance_ids):
    result = await self.conn.query_control(
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

  async def clear_pending_batches(self, model_id):
    result = await self.conn.query_control(
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
