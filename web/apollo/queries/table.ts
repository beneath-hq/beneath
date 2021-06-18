import gql from "graphql-tag";

export const QUERY_TABLE = gql`
  query TableByOrganizationProjectAndName($organizationName: String!, $projectName: String!, $tableName: String!) {
    tableByOrganizationProjectAndName(
      organizationName: $organizationName
      projectName: $projectName
      tableName: $tableName
    ) {
      tableID
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
        permissions {
          view
          create
          admin
        }
      }
      schemaKind
      schema
      avroSchema
      tableIndexes {
        indexID
        fields
        primary
        normalize
      }
      meta
      allowManualWrites
      useLog
      useIndex
      useWarehouse
      logRetentionSeconds
      indexRetentionSeconds
      warehouseRetentionSeconds
      primaryTableInstanceID
      primaryTableInstance {
        tableInstanceID
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

export const QUERY_TABLE_INSTANCE = gql`
  query TableInstanceByOrganizationProjectTableAndVersion(
    $organizationName: String!
    $projectName: String!
    $tableName: String!
    $version: Int!
  ) {
    tableInstanceByOrganizationProjectTableAndVersion(
      organizationName: $organizationName
      projectName: $projectName
      tableName: $tableName
      version: $version
    ) {
      tableInstanceID
      table {
        tableID
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
          permissions {
            view
            create
            admin
          }
        }
        schemaKind
        schema
        avroSchema
        tableIndexes {
          indexID
          fields
          primary
          normalize
        }
        meta
        allowManualWrites
        useLog
        useIndex
        useWarehouse
        logRetentionSeconds
        indexRetentionSeconds
        warehouseRetentionSeconds
        primaryTableInstanceID
        primaryTableInstance {
          tableInstanceID
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
      tableID
      version
      createdOn
      madePrimaryOn
      madeFinalOn
    }
  }
`;

export const QUERY_TABLE_INSTANCES = gql`
  query TableInstancesByOrganizationProjectAndTableName(
    $organizationName: String!
    $projectName: String!
    $tableName: String!
  ) {
    tableInstancesByOrganizationProjectAndTableName(
      organizationName: $organizationName
      projectName: $projectName
      tableName: $tableName
    ) {
      tableInstanceID
      tableID
      version
      createdOn
      madePrimaryOn
      madeFinalOn
    }
  }
`;

export const QUERY_TABLES_FOR_USER = gql`
  query TablesForUser($userID: UUID!) {
    tablesForUser(userID: $userID) {
      tableID
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
`;

export const COMPILE_SCHEMA = gql`
  query CompileSchema($input: CompileSchemaInput!) {
    compileSchema(input: $input) {
      canonicalIndexes
    }
  }
`;

export const CREATE_TABLE = gql`
  mutation CreateTable($input: CreateTableInput!) {
    createTable(input: $input) {
      tableID
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
      tableIndexes {
        indexID
        fields
        primary
        normalize
      }
      meta
      allowManualWrites
      useLog
      useIndex
      useWarehouse
      logRetentionSeconds
      indexRetentionSeconds
      warehouseRetentionSeconds
      primaryTableInstanceID
      instancesCreatedCount
      instancesDeletedCount
      instancesMadeFinalCount
      instancesMadePrimaryCount
    }
  }
`;

export const CREATE_TABLE_INSTANCE = gql`
  mutation CreateTableInstance($input: CreateTableInstanceInput!) {
    createTableInstance(input: $input) {
      tableInstanceID
      tableID
      version
      createdOn
      madePrimaryOn
      madeFinalOn
    }
  }
`;

export const UPDATE_TABLE_INSTANCE = gql`
  mutation UpdateTableInstance($input: UpdateTableInstanceInput!) {
    updateTableInstance(input: $input) {
      tableInstanceID
      tableID
      version
      createdOn
      madePrimaryOn
      madeFinalOn
    }
  }
`;

export const DELETE_TABLE_INSTANCE = gql`
  mutation DeleteTableInstance($instanceID: UUID!) {
    deleteTableInstance(instanceID: $instanceID)
  }
`;
