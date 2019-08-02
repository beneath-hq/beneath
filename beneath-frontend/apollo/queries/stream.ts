import gql from "graphql-tag";

export const QUERY_STREAM = gql`
  query QueryStream($name: String!, $projectName: String!) {
    stream(name: $name, projectName: $projectName) {
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
`;

export const CREATE_EXTERNAL_STREAM = gql`
  mutation CreateExternalStream(
    $projectID: UUID!,
    $description: String,
    $schema: String!,
    $batch: Boolean!,
    $manual: Boolean!
  ) {
    createExternalStream(
      projectID: $projectID,
      description: $description,
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
`;

export const UPDATE_STREAM = gql`
  mutation UpdateStream($streamID: UUID!, $description: String, $manual: Boolean) {
    updateStream(streamID: $streamID, description: $description, manual: $manual) {
      streamID
      description
      manual
    }
  }
`;
