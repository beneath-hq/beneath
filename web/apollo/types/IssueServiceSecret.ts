/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: IssueServiceSecret
// ====================================================

export interface IssueServiceSecret_issueServiceSecret_secret {
  __typename: "ServiceSecret";
  serviceSecretID: string;
  description: string;
  prefix: string;
  createdOn: ControlTime;
}

export interface IssueServiceSecret_issueServiceSecret {
  __typename: "NewServiceSecret";
  token: string;
  secret: IssueServiceSecret_issueServiceSecret_secret;
}

export interface IssueServiceSecret {
  issueServiceSecret: IssueServiceSecret_issueServiceSecret;
}

export interface IssueServiceSecretVariables {
  serviceID: ControlUUID;
  description: string;
}
