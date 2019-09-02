/* tslint:disable */
/* eslint-disable */
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: SecretsForProject
// ====================================================

export interface SecretsForProject_secretsForProject {
  __typename: "Secret";
  secretID: string;
  description: string;
  prefix: string;
  role: string;
  createdOn: ControlTime;
}

export interface SecretsForProject {
  secretsForProject: SecretsForProject_secretsForProject[];
}

export interface SecretsForProjectVariables {
  projectID: ControlUUID;
}
