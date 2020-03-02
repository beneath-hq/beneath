import gql from "graphql-tag";

export const QUERY_STREAM = gql`
  query QueryStream($name: String!, $projectName: String!) {
    streamByProjectAndName(name: $name, projectName: $projectName) {
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
      instancesCreatedCount
      instancesCommittedCount
      createdOn
      updatedOn
    }
  }
`;
