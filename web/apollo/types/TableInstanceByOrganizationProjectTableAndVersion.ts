/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { TableSchemaKind } from "./globalTypes";

// ====================================================
// GraphQL query operation: TableInstanceByOrganizationProjectTableAndVersion
// ====================================================

export interface TableInstanceByOrganizationProjectTableAndVersion_tableInstanceByOrganizationProjectTableAndVersion_table_project_organization {
  __typename: "PublicOrganization";
  organizationID: string;
  name: string;
}

export interface TableInstanceByOrganizationProjectTableAndVersion_tableInstanceByOrganizationProjectTableAndVersion_table_project_permissions {
  __typename: "PermissionsUsersProjects";
  view: boolean;
  create: boolean;
  admin: boolean;
}

export interface TableInstanceByOrganizationProjectTableAndVersion_tableInstanceByOrganizationProjectTableAndVersion_table_project {
  __typename: "Project";
  projectID: string;
  name: string;
  public: boolean;
  organization: TableInstanceByOrganizationProjectTableAndVersion_tableInstanceByOrganizationProjectTableAndVersion_table_project_organization;
  permissions: TableInstanceByOrganizationProjectTableAndVersion_tableInstanceByOrganizationProjectTableAndVersion_table_project_permissions;
}

export interface TableInstanceByOrganizationProjectTableAndVersion_tableInstanceByOrganizationProjectTableAndVersion_table_tableIndexes {
  __typename: "TableIndex";
  indexID: string;
  fields: string[];
  primary: boolean;
  normalize: boolean;
}

export interface TableInstanceByOrganizationProjectTableAndVersion_tableInstanceByOrganizationProjectTableAndVersion_table_primaryTableInstance {
  __typename: "TableInstance";
  tableInstanceID: string;
  createdOn: ControlTime;
  version: number;
  madePrimaryOn: ControlTime | null;
  madeFinalOn: ControlTime | null;
}

export interface TableInstanceByOrganizationProjectTableAndVersion_tableInstanceByOrganizationProjectTableAndVersion_table {
  __typename: "Table";
  tableID: string;
  name: string;
  description: string | null;
  createdOn: ControlTime;
  updatedOn: ControlTime;
  project: TableInstanceByOrganizationProjectTableAndVersion_tableInstanceByOrganizationProjectTableAndVersion_table_project;
  schemaKind: TableSchemaKind;
  schema: string;
  avroSchema: string;
  tableIndexes: TableInstanceByOrganizationProjectTableAndVersion_tableInstanceByOrganizationProjectTableAndVersion_table_tableIndexes[];
  meta: boolean;
  allowManualWrites: boolean;
  useLog: boolean;
  useIndex: boolean;
  useWarehouse: boolean;
  logRetentionSeconds: number;
  indexRetentionSeconds: number;
  warehouseRetentionSeconds: number;
  primaryTableInstanceID: ControlUUID | null;
  primaryTableInstance: TableInstanceByOrganizationProjectTableAndVersion_tableInstanceByOrganizationProjectTableAndVersion_table_primaryTableInstance | null;
  instancesCreatedCount: number;
  instancesDeletedCount: number;
  instancesMadeFinalCount: number;
  instancesMadePrimaryCount: number;
}

export interface TableInstanceByOrganizationProjectTableAndVersion_tableInstanceByOrganizationProjectTableAndVersion {
  __typename: "TableInstance";
  tableInstanceID: string;
  table: TableInstanceByOrganizationProjectTableAndVersion_tableInstanceByOrganizationProjectTableAndVersion_table;
  tableID: ControlUUID;
  version: number;
  createdOn: ControlTime;
  madePrimaryOn: ControlTime | null;
  madeFinalOn: ControlTime | null;
}

export interface TableInstanceByOrganizationProjectTableAndVersion {
  tableInstanceByOrganizationProjectTableAndVersion: TableInstanceByOrganizationProjectTableAndVersion_tableInstanceByOrganizationProjectTableAndVersion;
}

export interface TableInstanceByOrganizationProjectTableAndVersionVariables {
  organizationName: string;
  projectName: string;
  tableName: string;
  version: number;
}
