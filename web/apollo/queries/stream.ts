import gql from "graphql-tag";

export const QUERY_STREAM = gql`
  query StreamByOrganizationProjectAndName($organizationName: String!, $projectName: String!, $streamName: String!) {
    streamByOrganizationProjectAndName(organizationName: $organizationName, projectName: $projectName, streamName: $streamName) {
      streamID
      name
      description
      createdOn
      updatedOn
      project {
        projectID
        name
        organization {
          organizationID
          name
        }
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
