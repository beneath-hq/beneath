/* tslint:disable */
/* eslint-disable */
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: QueryStream
// ====================================================

export interface QueryStream_stream_project {
  __typename: "Project";
  projectID: string;
  name: string;
}

export interface QueryStream_stream {
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
  project: QueryStream_stream_project;
  currentStreamInstanceID: ControlUUID | null;
  instancesCreatedCount: number;
  instancesCommittedCount: number;
  createdOn: ControlTime;
  updatedOn: ControlTime;
}

export interface QueryStream {
  stream: QueryStream_stream;
}

export interface QueryStreamVariables {
  name: string;
  projectName: string;
}
