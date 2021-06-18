/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { CreateTableInput, TableSchemaKind } from "./globalTypes";

// ====================================================
// GraphQL mutation operation: CreateTable
// ====================================================

export interface CreateTable_createTable_project_organization {
  __typename: "PublicOrganization";
  organizationID: string;
  name: string;
}

export interface CreateTable_createTable_project {
  __typename: "Project";
  projectID: string;
  name: string;
  organization: CreateTable_createTable_project_organization;
}

export interface CreateTable_createTable_tableIndexes {
  __typename: "TableIndex";
  indexID: string;
  fields: string[];
  primary: boolean;
  normalize: boolean;
}

export interface CreateTable_createTable {
  __typename: "Table";
  tableID: string;
  name: string;
  description: string | null;
  createdOn: ControlTime;
  updatedOn: ControlTime;
  project: CreateTable_createTable_project;
  schemaKind: TableSchemaKind;
  schema: string;
  avroSchema: string;
  tableIndexes: CreateTable_createTable_tableIndexes[];
  meta: boolean;
  allowManualWrites: boolean;
  useLog: boolean;
  useIndex: boolean;
  useWarehouse: boolean;
  logRetentionSeconds: number;
  indexRetentionSeconds: number;
  warehouseRetentionSeconds: number;
  primaryTableInstanceID: ControlUUID | null;
  instancesCreatedCount: number;
  instancesDeletedCount: number;
  instancesMadeFinalCount: number;
  instancesMadePrimaryCount: number;
}

export interface CreateTable {
  createTable: CreateTable_createTable;
}

export interface CreateTableVariables {
  input: CreateTableInput;
}
