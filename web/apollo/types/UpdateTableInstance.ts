/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { UpdateTableInstanceInput } from "./globalTypes";

// ====================================================
// GraphQL mutation operation: UpdateTableInstance
// ====================================================

export interface UpdateTableInstance_updateTableInstance {
  __typename: "TableInstance";
  tableInstanceID: string;
  tableID: ControlUUID;
  version: number;
  createdOn: ControlTime;
  madePrimaryOn: ControlTime | null;
  madeFinalOn: ControlTime | null;
}

export interface UpdateTableInstance {
  updateTableInstance: UpdateTableInstance_updateTableInstance;
}

export interface UpdateTableInstanceVariables {
  input: UpdateTableInstanceInput;
}
