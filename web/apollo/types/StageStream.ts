/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { StreamSchemaKind } from "./globalTypes";

// ====================================================
// GraphQL mutation operation: StageStream
// ====================================================

export interface StageStream_stageStream_project_organization {
  __typename: "PublicOrganization";
  organizationID: string;
  name: string;
}

export interface StageStream_stageStream_project {
  __typename: "Project";
  projectID: string;
  name: string;
  organization: StageStream_stageStream_project_organization;
}

export interface StageStream_stageStream_streamIndexes {
  __typename: "StreamIndex";
  indexID: string;
  fields: string[];
  primary: boolean;
  normalize: boolean;
}

export interface StageStream_stageStream {
  __typename: "Stream";
  streamID: string;
  name: string;
  description: string | null;
  createdOn: ControlTime;
  updatedOn: ControlTime;
  project: StageStream_stageStream_project;
  schemaKind: StreamSchemaKind;
  schema: string;
  avroSchema: string;
  streamIndexes: StageStream_stageStream_streamIndexes[];
  allowManualWrites: boolean;
  useLog: boolean;
  useIndex: boolean;
  useWarehouse: boolean;
  logRetentionSeconds: number;
  indexRetentionSeconds: number;
  warehouseRetentionSeconds: number;
  primaryStreamInstanceID: ControlUUID | null;
  instancesCreatedCount: number;
  instancesDeletedCount: number;
  instancesMadeFinalCount: number;
  instancesMadePrimaryCount: number;
}

export interface StageStream {
  stageStream: StageStream_stageStream;
}

export interface StageStreamVariables {
  organizationName: string;
  projectName: string;
  streamName: string;
  schemaKind: StreamSchemaKind;
  schema: string;
  indexes?: string | null;
  description?: string | null;
  allowManualWrites?: boolean | null;
  useLog?: boolean | null;
  useIndex?: boolean | null;
  useWarehouse?: boolean | null;
  logRetentionSeconds?: number | null;
  indexRetentionSeconds?: number | null;
  warehouseRetentionSeconds?: number | null;
}
