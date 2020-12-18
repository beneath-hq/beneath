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

export const STAGE_STREAM = gql`
  mutation StageStream(
    $organizationName: String!,
    $projectName: String!,
    $streamName: String!
    $schemaKind: StreamSchemaKind!,
    $schema: String!,
    $indexes: String,
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
      indexes: $indexes,
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

export const STAGE_STREAM_INSTANCE = gql`
  mutation StageStreamInstance($streamID: UUID!, $version: Int!, $makeFinal: Boolean, $makePrimary: Boolean){
  stageStreamInstance(
    streamID: $streamID,
    version: $version,
    makeFinal: $makeFinal,
    makePrimary: $makePrimary,
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

export const DELETE_STREAM_INSTANCE = gql`
  mutation DeleteStreamInstance($instanceID: UUID!){
  deleteStreamInstance(
    instanceID: $instanceID,
  )
}
`;
