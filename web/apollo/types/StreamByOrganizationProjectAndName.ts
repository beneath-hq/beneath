/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { StreamSchemaKind } from "./globalTypes";

// ====================================================
// GraphQL query operation: StreamByOrganizationProjectAndName
// ====================================================

export interface StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName_project_organization {
  __typename: "PublicOrganization";
  organizationID: string;
  name: string;
}

export interface StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName_project_permissions {
  __typename: "PermissionsUsersProjects";
  view: boolean;
  create: boolean;
  admin: boolean;
}

export interface StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName_project {
  __typename: "Project";
  projectID: string;
  name: string;
  public: boolean;
  organization: StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName_project_organization;
  permissions: StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName_project_permissions;
}

export interface StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName_streamIndexes {
  __typename: "StreamIndex";
  indexID: string;
  fields: string[];
  primary: boolean;
  normalize: boolean;
}

export interface StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName_primaryStreamInstance {
  __typename: "StreamInstance";
  streamInstanceID: string;
  createdOn: ControlTime;
  version: number;
  madePrimaryOn: ControlTime | null;
  madeFinalOn: ControlTime | null;
}

export interface StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName {
  __typename: "Stream";
  streamID: string;
  name: string;
  description: string | null;
  createdOn: ControlTime;
  updatedOn: ControlTime;
  project: StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName_project;
  schemaKind: StreamSchemaKind;
  schema: string;
  avroSchema: string;
  streamIndexes: StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName_streamIndexes[];
  meta: boolean;
  allowManualWrites: boolean;
  useLog: boolean;
  useIndex: boolean;
  useWarehouse: boolean;
  logRetentionSeconds: number;
  indexRetentionSeconds: number;
  warehouseRetentionSeconds: number;
  primaryStreamInstanceID: ControlUUID | null;
  primaryStreamInstance: StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName_primaryStreamInstance | null;
  instancesCreatedCount: number;
  instancesDeletedCount: number;
  instancesMadeFinalCount: number;
  instancesMadePrimaryCount: number;
}

export interface StreamByOrganizationProjectAndName {
  streamByOrganizationProjectAndName: StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName;
}

export interface StreamByOrganizationProjectAndNameVariables {
  organizationName: string;
  projectName: string;
  streamName: string;
}
