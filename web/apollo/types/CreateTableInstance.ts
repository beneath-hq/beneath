/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { CreateTableInstanceInput } from "./globalTypes";

// ====================================================
// GraphQL mutation operation: CreateTableInstance
// ====================================================

export interface CreateTableInstance_createTableInstance {
  __typename: "TableInstance";
  tableInstanceID: string;
  tableID: ControlUUID;
  version: number;
  createdOn: ControlTime;
  madePrimaryOn: ControlTime | null;
  madeFinalOn: ControlTime | null;
}

export interface CreateTableInstance {
  createTableInstance: CreateTableInstance_createTableInstance;
}

export interface CreateTableInstanceVariables {
  input: CreateTableInstanceInput;
}
