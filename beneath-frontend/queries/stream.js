import gql from "graphql-tag";

export const QUERY_STREAM = gql`
  query QueryStream($name: String!, $projectName: String!) {
    stream(name: $name, projectName: $projectName) {
      streamId
      name
      description
      schema
      schemaType
      batch
      external
      manual
      project {
        projectId
        name
      }
      createdOn
      updatedOn
    }
  }
`;

export const CREATE_EXTERNAL_STREAM = gql`
  mutation CreateExternalStream(
    $projectId: ID!,
    $name: String!,
    $description: String!,
    $avroSchema: String!,
    $batch: Boolean!,
    $manual: Boolean!
  ) {
    createExternalStream(
      projectId: $projectId,
      name: $name,
      description: $description,
      avroSchema: $avroSchema,
      batch: $batch,
      manual: $manual
    ) {
      streamId
      name
      description
      schema
      schemaType
      batch
      external
      manual
      project {
        projectId
        name
      }
      createdOn
      updatedOn
    }
  }
`;

export const UPDATE_STREAM = gql`
  mutation UpdateStream($streamId: ID!, $description: String!, $manual: Boolean!) {
    updateStream(streamId: $streamId, description: $description, manual: $manual) {
      streamId
      description
      manual
    }
  }
`;
