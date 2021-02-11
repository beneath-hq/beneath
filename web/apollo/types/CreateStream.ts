/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { CreateStreamInput, StreamSchemaKind } from "./globalTypes";

// ====================================================
// GraphQL mutation operation: CreateStream
// ====================================================

export interface CreateStream_createStream_project_organization {
  __typename: "PublicOrganization";
  organizationID: string;
  name: string;
}

export interface CreateStream_createStream_project {
  __typename: "Project";
  projectID: string;
  name: string;
  organization: CreateStream_createStream_project_organization;
}

export interface CreateStream_createStream_streamIndexes {
  __typename: "StreamIndex";
  indexID: string;
  fields: string[];
  primary: boolean;
  normalize: boolean;
}

export interface CreateStream_createStream {
  __typename: "Stream";
  streamID: string;
  name: string;
  description: string | null;
  createdOn: ControlTime;
  updatedOn: ControlTime;
  project: CreateStream_createStream_project;
  schemaKind: StreamSchemaKind;
  schema: string;
  avroSchema: string;
  streamIndexes: CreateStream_createStream_streamIndexes[];
  meta: boolean;
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

export interface CreateStream {
  createStream: CreateStream_createStream;
}

export interface CreateStreamVariables {
  input: CreateStreamInput;
}
