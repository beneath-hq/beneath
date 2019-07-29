/* tslint:disable */
/* eslint-disable */
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: IssueProjectKey
// ====================================================

export interface IssueProjectKey_issueProjectKey_key {
  __typename: "Key";
  keyID: string;
  description: string;
  prefix: string;
  role: string;
  createdOn: ControlTime;
}

export interface IssueProjectKey_issueProjectKey {
  __typename: "NewKey";
  keyString: string;
  key: IssueProjectKey_issueProjectKey_key;
}

export interface IssueProjectKey {
  issueProjectKey: IssueProjectKey_issueProjectKey;
}

export interface IssueProjectKeyVariables {
  projectID: ControlUUID;
  readonly: boolean;
  description: string;
}
