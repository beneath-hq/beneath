/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: CreateRecords
// ====================================================

export interface CreateRecords_createRecords {
  __typename: "CreateRecordsResponse";
  error: string | null;
}

export interface CreateRecords {
  createRecords: CreateRecords_createRecords;
}

export interface CreateRecordsVariables {
  instanceID: ControlUUID;
  json: ControlJSON;
}
