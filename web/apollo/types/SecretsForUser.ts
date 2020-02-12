/* tslint:disable */
/* eslint-disable */
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: SecretsForUser
// ====================================================

export interface SecretsForUser_secretsForUser {
  __typename: "UserSecret";
  userSecretID: string;
  description: string;
  prefix: string;
  createdOn: ControlTime;
  readOnly: boolean;
  publicOnly: boolean;
}

export interface SecretsForUser {
  secretsForUser: SecretsForUser_secretsForUser[];
}

export interface SecretsForUserVariables {
  userID: ControlUUID;
}
