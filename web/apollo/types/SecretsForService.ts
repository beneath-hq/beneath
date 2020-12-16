/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: SecretsForService
// ====================================================

export interface SecretsForService_secretsForService {
  __typename: "ServiceSecret";
  serviceSecretID: string;
  description: string;
  prefix: string;
  createdOn: ControlTime;
}

export interface SecretsForService {
  secretsForService: SecretsForService_secretsForService[];
}

export interface SecretsForServiceVariables {
  serviceID: ControlUUID;
}
