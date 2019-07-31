/* tslint:disable */
/* eslint-disable */
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: KeysForUser
// ====================================================

export interface KeysForUser_keysForUser {
  __typename: "Key";
  keyID: string;
  description: string;
  prefix: string;
  role: string;
  createdOn: ControlTime;
}

export interface KeysForUser {
  keysForUser: KeysForUser_keysForUser[];
}

export interface KeysForUserVariables {
  userID: ControlUUID;
}
