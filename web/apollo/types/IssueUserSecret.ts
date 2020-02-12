/* tslint:disable */
/* eslint-disable */
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: IssueUserSecret
// ====================================================

export interface IssueUserSecret_issueUserSecret_secret {
  __typename: "UserSecret";
  userSecretID: string;
  description: string;
  prefix: string;
  createdOn: ControlTime;
  readOnly: boolean;
  publicOnly: boolean;
}

export interface IssueUserSecret_issueUserSecret {
  __typename: "NewUserSecret";
  token: string;
  secret: IssueUserSecret_issueUserSecret_secret;
}

export interface IssueUserSecret {
  issueUserSecret: IssueUserSecret_issueUserSecret;
}

export interface IssueUserSecretVariables {
  description: string;
  readOnly: boolean;
  publicOnly: boolean;
}
