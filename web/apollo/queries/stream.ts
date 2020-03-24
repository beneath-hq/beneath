import gql from "graphql-tag";

export const QUERY_STREAM = gql`
  query QueryStream($name: String!, $projectName: String!) {
    streamByProjectAndName(name: $name, projectName: $projectName) {
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
        indexID
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
`;
