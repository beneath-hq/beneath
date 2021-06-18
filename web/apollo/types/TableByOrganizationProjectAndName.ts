/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { TableSchemaKind } from "./globalTypes";

// ====================================================
// GraphQL query operation: TableByOrganizationProjectAndName
// ====================================================

export interface TableByOrganizationProjectAndName_tableByOrganizationProjectAndName_project_organization {
  __typename: "PublicOrganization";
  organizationID: string;
  name: string;
}

export interface TableByOrganizationProjectAndName_tableByOrganizationProjectAndName_project_permissions {
  __typename: "PermissionsUsersProjects";
  view: boolean;
  create: boolean;
  admin: boolean;
}

export interface TableByOrganizationProjectAndName_tableByOrganizationProjectAndName_project {
  __typename: "Project";
  projectID: string;
  name: string;
  public: boolean;
  organization: TableByOrganizationProjectAndName_tableByOrganizationProjectAndName_project_organization;
  permissions: TableByOrganizationProjectAndName_tableByOrganizationProjectAndName_project_permissions;
}

export interface TableByOrganizationProjectAndName_tableByOrganizationProjectAndName_tableIndexes {
  __typename: "TableIndex";
  indexID: string;
  fields: string[];
  primary: boolean;
  normalize: boolean;
}

export interface TableByOrganizationProjectAndName_tableByOrganizationProjectAndName_primaryTableInstance {
  __typename: "TableInstance";
  tableInstanceID: string;
  createdOn: ControlTime;
  version: number;
  madePrimaryOn: ControlTime | null;
  madeFinalOn: ControlTime | null;
}

export interface TableByOrganizationProjectAndName_tableByOrganizationProjectAndName {
  __typename: "Table";
  tableID: string;
  name: string;
  description: string | null;
  createdOn: ControlTime;
  updatedOn: ControlTime;
  project: TableByOrganizationProjectAndName_tableByOrganizationProjectAndName_project;
  schemaKind: TableSchemaKind;
  schema: string;
  avroSchema: string;
  tableIndexes: TableByOrganizationProjectAndName_tableByOrganizationProjectAndName_tableIndexes[];
  meta: boolean;
  allowManualWrites: boolean;
  useLog: boolean;
  useIndex: boolean;
  useWarehouse: boolean;
  logRetentionSeconds: number;
  indexRetentionSeconds: number;
  warehouseRetentionSeconds: number;
  primaryTableInstanceID: ControlUUID | null;
  primaryTableInstance: TableByOrganizationProjectAndName_tableByOrganizationProjectAndName_primaryTableInstance | null;
  instancesCreatedCount: number;
  instancesDeletedCount: number;
  instancesMadeFinalCount: number;
  instancesMadePrimaryCount: number;
}

export interface TableByOrganizationProjectAndName {
  tableByOrganizationProjectAndName: TableByOrganizationProjectAndName_tableByOrganizationProjectAndName;
}

export interface TableByOrganizationProjectAndNameVariables {
  organizationName: string;
  projectName: string;
  tableName: string;
}
