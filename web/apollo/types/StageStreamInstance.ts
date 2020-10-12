/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: StageStreamInstance
// ====================================================

export interface StageStreamInstance_stageStreamInstance {
  __typename: "StreamInstance";
  streamInstanceID: string;
  streamID: ControlUUID;
  version: number;
  createdOn: ControlTime;
  madePrimaryOn: ControlTime | null;
  madeFinalOn: ControlTime | null;
}

export interface StageStreamInstance {
  stageStreamInstance: StageStreamInstance_stageStreamInstance;
}

export interface StageStreamInstanceVariables {
  streamID: ControlUUID;
  version: number;
  makeFinal?: boolean | null;
  makePrimary?: boolean | null;
}
