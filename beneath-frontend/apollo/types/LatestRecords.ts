/* tslint:disable */
/* eslint-disable */
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: LatestRecords
// ====================================================

export interface LatestRecords_latestRecords_data {
  __typename: "Record";
  recordID: string;
  data: ControlJSON;
  sequenceNumber: string;
}

export interface LatestRecords_latestRecords {
  __typename: "RecordsResponse";
  data: LatestRecords_latestRecords_data[] | null;
  error: string | null;
}

export interface LatestRecords {
  latestRecords: LatestRecords_latestRecords;
}

export interface LatestRecordsVariables {
  projectName: string;
  streamName: string;
  limit: number;
  before?: string | null;
}
