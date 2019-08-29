/* tslint:disable */
/* eslint-disable */
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: IssueProjectSecret
// ====================================================

export interface IssueProjectSecret_issueProjectSecret_secret {
  __typename: "Secret";
  secretID: string;
  description: string;
  prefix: string;
  role: string;
  createdOn: ControlTime;
}

export interface IssueProjectSecret_issueProjectSecret {
  __typename: "NewSecret";
  secretString: string;
  secret: IssueProjectSecret_issueProjectSecret_secret;
}

export interface IssueProjectSecret {
  issueProjectSecret: IssueProjectSecret_issueProjectSecret;
}

export interface IssueProjectSecretVariables {
  projectID: ControlUUID;
  readonly: boolean;
  description: string;
}
