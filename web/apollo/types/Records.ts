/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: Records
// ====================================================

export interface Records_records_data {
  __typename: "Record";
  recordID: string;
  data: ControlJSON;
  timestamp: number;
}

export interface Records_records {
  __typename: "RecordsResponse";
  data: Records_records_data[] | null;
  error: string | null;
}

export interface Records {
  records: Records_records;
}

export interface RecordsVariables {
  projectName: string;
  streamName: string;
  limit: number;
  where?: ControlJSON | null;
  after?: ControlJSON | null;
}
