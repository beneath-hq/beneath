/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: TablesForUser
// ====================================================

export interface TablesForUser_tablesForUser_project_organization {
  __typename: "PublicOrganization";
  organizationID: string;
  name: string;
}

export interface TablesForUser_tablesForUser_project {
  __typename: "Project";
  projectID: string;
  name: string;
  public: boolean;
  organization: TablesForUser_tablesForUser_project_organization;
}

export interface TablesForUser_tablesForUser {
  __typename: "Table";
  tableID: string;
  name: string;
  description: string | null;
  createdOn: ControlTime;
  updatedOn: ControlTime;
  project: TablesForUser_tablesForUser_project;
}

export interface TablesForUser {
  tablesForUser: TablesForUser_tablesForUser[];
}

export interface TablesForUserVariables {
  userID: ControlUUID;
}
