/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { UpdateStreamInstanceInput } from "./globalTypes";

// ====================================================
// GraphQL mutation operation: UpdateStreamInstance
// ====================================================

export interface UpdateStreamInstance_updateStreamInstance {
  __typename: "StreamInstance";
  streamInstanceID: string;
  streamID: ControlUUID;
  version: number;
  createdOn: ControlTime;
  madePrimaryOn: ControlTime | null;
  madeFinalOn: ControlTime | null;
}

export interface UpdateStreamInstance {
  updateStreamInstance: UpdateStreamInstance_updateStreamInstance;
}

export interface UpdateStreamInstanceVariables {
  input: UpdateStreamInstanceInput;
}
