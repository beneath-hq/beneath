/* tslint:disable */
/* eslint-disable */
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: IssueUserKey
// ====================================================

export interface IssueUserKey_issueUserKey_key {
  __typename: "Key";
  keyID: string;
  description: string;
  prefix: string;
  role: string;
  createdOn: ControlTime;
}

export interface IssueUserKey_issueUserKey {
  __typename: "NewKey";
  keyString: string;
  key: IssueUserKey_issueUserKey_key;
}

export interface IssueUserKey {
  issueUserKey: IssueUserKey_issueUserKey;
}

export interface IssueUserKeyVariables {
  readonly: boolean;
  description: string;
}
