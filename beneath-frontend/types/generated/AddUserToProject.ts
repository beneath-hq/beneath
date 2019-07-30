/* tslint:disable */
/* eslint-disable */
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: AddUserToProject
// ====================================================

export interface AddUserToProject_addUserToProject {
  __typename: "User";
  userID: string;
  name: string;
  username: string | null;
  photoURL: string | null;
}

export interface AddUserToProject {
  addUserToProject: AddUserToProject_addUserToProject | null;
}

export interface AddUserToProjectVariables {
  email: string;
  projectID: ControlUUID;
}
