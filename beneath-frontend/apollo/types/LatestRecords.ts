/* tslint:disable */
/* eslint-disable */
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: LatestRecords
// ====================================================

export interface LatestRecords_latestRecords {
  __typename: "Record";
  recordID: string;
  data: ControlJSON;
  sequenceNumber: string;
}

export interface LatestRecords {
  latestRecords: LatestRecords_latestRecords[];
}

export interface LatestRecordsVariables {
  projectName: string;
  streamName: string;
  limit: number;
}
