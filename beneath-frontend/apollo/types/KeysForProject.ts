/* tslint:disable */
/* eslint-disable */
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: KeysForProject
// ====================================================

export interface KeysForProject_keysForProject {
  __typename: "Key";
  keyID: string;
  description: string;
  prefix: string;
  role: string;
  createdOn: ControlTime;
}

export interface KeysForProject {
  keysForProject: KeysForProject_keysForProject[];
}

export interface KeysForProjectVariables {
  projectID: ControlUUID;
}
