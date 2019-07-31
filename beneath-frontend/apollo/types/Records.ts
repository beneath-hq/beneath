/* tslint:disable */
/* eslint-disable */
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: Records
// ====================================================

export interface Records_records {
  __typename: "Record";
  recordID: string;
  data: ControlJSON;
  sequenceNumber: string;
}

export interface Records {
  records: Records_records[];
}

export interface RecordsVariables {
  projectName: string;
  streamName: string;
  keyFields: string[];
  limit: number;
  where?: ControlJSON | null;
}
