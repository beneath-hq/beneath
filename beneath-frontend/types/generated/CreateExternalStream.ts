/* tslint:disable */
/* eslint-disable */
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: CreateExternalStream
// ====================================================

export interface CreateExternalStream_createExternalStream_project {
  __typename: "Project";
  projectID: string;
  name: string;
}

export interface CreateExternalStream_createExternalStream {
  __typename: "Stream";
  streamID: string;
  name: string;
  description: string | null;
  schema: string;
  avroSchema: string;
  keyFields: string[];
  external: boolean;
  batch: boolean;
  manual: boolean;
  project: CreateExternalStream_createExternalStream_project;
  createdOn: ControlTime;
  updatedOn: ControlTime;
}

export interface CreateExternalStream {
  createExternalStream: CreateExternalStream_createExternalStream;
}

export interface CreateExternalStreamVariables {
  projectID: ControlUUID;
  description?: string | null;
  schema: string;
  batch: boolean;
  manual: boolean;
}
