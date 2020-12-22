/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { CreateStreamInstanceInput } from "./globalTypes";

// ====================================================
// GraphQL mutation operation: CreateStreamInstance
// ====================================================

export interface CreateStreamInstance_createStreamInstance {
  __typename: "StreamInstance";
  streamInstanceID: string;
  streamID: ControlUUID;
  version: number;
  createdOn: ControlTime;
  madePrimaryOn: ControlTime | null;
  madeFinalOn: ControlTime | null;
}

export interface CreateStreamInstance {
  createStreamInstance: CreateStreamInstance_createStreamInstance;
}

export interface CreateStreamInstanceVariables {
  input: CreateStreamInstanceInput;
}
