/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: QueryStream
// ====================================================

export interface QueryStream_streamByProjectAndName_project {
  __typename: "Project";
  projectID: string;
  name: string;
}

export interface QueryStream_streamByProjectAndName_streamIndexes {
  __typename: "StreamIndex";
  indexID: string;
  fields: string[];
  primary: boolean;
  normalize: boolean;
}

export interface QueryStream_streamByProjectAndName {
  __typename: "Stream";
  streamID: string;
  name: string;
  description: string | null;
  createdOn: ControlTime;
  updatedOn: ControlTime;
  project: QueryStream_streamByProjectAndName_project;
  schema: string;
  avroSchema: string;
  streamIndexes: QueryStream_streamByProjectAndName_streamIndexes[];
  external: boolean;
  batch: boolean;
  manual: boolean;
  retentionSeconds: number;
  instancesCreatedCount: number;
  instancesCommittedCount: number;
  currentStreamInstanceID: ControlUUID | null;
}

export interface QueryStream {
  streamByProjectAndName: QueryStream_streamByProjectAndName;
}

export interface QueryStreamVariables {
  name: string;
  projectName: string;
}
