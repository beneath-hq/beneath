/* tslint:disable */
/* eslint-disable */
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: IssueUserSecret
// ====================================================

export interface IssueUserSecret_issueUserSecret_secret {
  __typename: "Secret";
  secretID: string;
  description: string;
  prefix: string;
  role: string;
  createdOn: ControlTime;
}

export interface IssueUserSecret_issueUserSecret {
  __typename: "NewSecret";
  secretString: string;
  secret: IssueUserSecret_issueUserSecret_secret;
}

export interface IssueUserSecret {
  issueUserSecret: IssueUserSecret_issueUserSecret;
}

export interface IssueUserSecretVariables {
  readonly: boolean;
  description: string;
}
