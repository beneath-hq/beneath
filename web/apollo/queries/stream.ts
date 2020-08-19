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
      schemaKind
      schema
      avroSchema
      streamIndexes {
        indexID
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
`;

export const QUERY_STREAM_INSTANCES = gql`
  query StreamInstancesByOrganizationProjectAndStreamName($organizationName: String!, $projectName: String!, $streamName: String!) {
    streamInstancesByOrganizationProjectAndStreamName(
      organizationName: $organizationName, projectName: $projectName, streamName: $streamName
    ) {
      streamInstanceID
      streamID
      version
      createdOn
      madePrimaryOn
      madeFinalOn
    }
  }
`;