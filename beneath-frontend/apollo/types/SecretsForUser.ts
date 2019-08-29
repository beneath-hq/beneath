/* tslint:disable */
/* eslint-disable */
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: SecretsForUser
// ====================================================

export interface SecretsForUser_secretsForUser {
  __typename: "Secret";
  secretID: string;
  description: string;
  prefix: string;
  role: string;
  createdOn: ControlTime;
}

export interface SecretsForUser {
  secretsForUser: SecretsForUser_secretsForUser[];
}

export interface SecretsForUserVariables {
  userID: ControlUUID;
}
