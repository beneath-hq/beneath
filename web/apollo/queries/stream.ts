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
        public
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

export const QUERY_STREAMS_FOR_USER = gql`
  query StreamsForUser($userID: UUID!) {
    streamsForUser(userID: $userID) {
      streamID
      name
      description
      createdOn
      updatedOn
      project {
        projectID
        name
        public
        organization {
          organizationID
          name
        }
      }
    }
  }
`

export const COMPILE_SCHEMA = gql`
  query CompileSchema($input: CompileSchemaInput!) {
    compileSchema(input: $input) {
      canonicalIndexes
    }
  }
`;

export const CREATE_STREAM = gql`
  mutation CreateStream($input: CreateStreamInput!) {
    createStream(input: $input) {
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
      instancesCreatedCount
      instancesDeletedCount
      instancesMadeFinalCount
      instancesMadePrimaryCount
    }
  }
`;

export const CREATE_STREAM_INSTANCE = gql`
  mutation CreateStreamInstance($input: CreateStreamInstanceInput!) {
    createStreamInstance(input: $input) {
      streamInstanceID
      streamID
      version
      createdOn
      madePrimaryOn
      madeFinalOn
    }
  }
`;

export const UPDATE_STREAM_INSTANCE = gql`
  mutation UpdateStreamInstance($input: UpdateStreamInstanceInput!) {
    updateStreamInstance(input: $input) {
      streamInstanceID
      streamID
      version
      createdOn
      madePrimaryOn
      madeFinalOn
    }
  }
`;

export const DELETE_STREAM_INSTANCE = gql`
  mutation DeleteStreamInstance($instanceID: UUID!){
    deleteStreamInstance(
      instanceID: $instanceID,
    )
  }
`;
