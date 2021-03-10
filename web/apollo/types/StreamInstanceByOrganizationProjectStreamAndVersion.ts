/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { StreamSchemaKind } from "./globalTypes";

// ====================================================
// GraphQL query operation: StreamInstanceByOrganizationProjectStreamAndVersion
// ====================================================

export interface StreamInstanceByOrganizationProjectStreamAndVersion_streamInstanceByOrganizationProjectStreamAndVersion_stream_project_organization {
  __typename: "PublicOrganization";
  organizationID: string;
  name: string;
}

export interface StreamInstanceByOrganizationProjectStreamAndVersion_streamInstanceByOrganizationProjectStreamAndVersion_stream_project_permissions {
  __typename: "PermissionsUsersProjects";
  view: boolean;
  create: boolean;
  admin: boolean;
}

export interface StreamInstanceByOrganizationProjectStreamAndVersion_streamInstanceByOrganizationProjectStreamAndVersion_stream_project {
  __typename: "Project";
  projectID: string;
  name: string;
  public: boolean;
  organization: StreamInstanceByOrganizationProjectStreamAndVersion_streamInstanceByOrganizationProjectStreamAndVersion_stream_project_organization;
  permissions: StreamInstanceByOrganizationProjectStreamAndVersion_streamInstanceByOrganizationProjectStreamAndVersion_stream_project_permissions;
}

export interface StreamInstanceByOrganizationProjectStreamAndVersion_streamInstanceByOrganizationProjectStreamAndVersion_stream_streamIndexes {
  __typename: "StreamIndex";
  indexID: string;
  fields: string[];
  primary: boolean;
  normalize: boolean;
}

export interface StreamInstanceByOrganizationProjectStreamAndVersion_streamInstanceByOrganizationProjectStreamAndVersion_stream_primaryStreamInstance {
  __typename: "StreamInstance";
  streamInstanceID: string;
  createdOn: ControlTime;
  version: number;
  madePrimaryOn: ControlTime | null;
  madeFinalOn: ControlTime | null;
}

export interface StreamInstanceByOrganizationProjectStreamAndVersion_streamInstanceByOrganizationProjectStreamAndVersion_stream {
  __typename: "Stream";
  streamID: string;
  name: string;
  description: string | null;
  createdOn: ControlTime;
  updatedOn: ControlTime;
  project: StreamInstanceByOrganizationProjectStreamAndVersion_streamInstanceByOrganizationProjectStreamAndVersion_stream_project;
  schemaKind: StreamSchemaKind;
  schema: string;
  avroSchema: string;
  streamIndexes: StreamInstanceByOrganizationProjectStreamAndVersion_streamInstanceByOrganizationProjectStreamAndVersion_stream_streamIndexes[];
  meta: boolean;
  allowManualWrites: boolean;
  useLog: boolean;
  useIndex: boolean;
  useWarehouse: boolean;
  logRetentionSeconds: number;
  indexRetentionSeconds: number;
  warehouseRetentionSeconds: number;
  primaryStreamInstanceID: ControlUUID | null;
  primaryStreamInstance: StreamInstanceByOrganizationProjectStreamAndVersion_streamInstanceByOrganizationProjectStreamAndVersion_stream_primaryStreamInstance | null;
  instancesCreatedCount: number;
  instancesDeletedCount: number;
  instancesMadeFinalCount: number;
  instancesMadePrimaryCount: number;
}

export interface StreamInstanceByOrganizationProjectStreamAndVersion_streamInstanceByOrganizationProjectStreamAndVersion {
  __typename: "StreamInstance";
  streamInstanceID: string;
  stream: StreamInstanceByOrganizationProjectStreamAndVersion_streamInstanceByOrganizationProjectStreamAndVersion_stream;
  streamID: ControlUUID;
  version: number;
  createdOn: ControlTime;
  madePrimaryOn: ControlTime | null;
  madeFinalOn: ControlTime | null;
}

export interface StreamInstanceByOrganizationProjectStreamAndVersion {
  streamInstanceByOrganizationProjectStreamAndVersion: StreamInstanceByOrganizationProjectStreamAndVersion_streamInstanceByOrganizationProjectStreamAndVersion;
}

export interface StreamInstanceByOrganizationProjectStreamAndVersionVariables {
  organizationName: string;
  projectName: string;
  streamName: string;
  version: number;
}
