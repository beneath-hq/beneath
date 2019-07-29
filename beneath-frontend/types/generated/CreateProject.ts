/* tslint:disable */
/* eslint-disable */
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: CreateProject
// ====================================================

export interface CreateProject_createProject_users {
  __typename: "User";
  userID: string;
  name: string;
  username: string | null;
  photoURL: string | null;
}

export interface CreateProject_createProject {
  __typename: "Project";
  projectID: string;
  name: string;
  displayName: string;
  site: string | null;
  description: string | null;
  photoURL: string | null;
  createdOn: ControlTime;
  updatedOn: ControlTime;
  users: CreateProject_createProject_users[];
}

export interface CreateProject {
  createProject: CreateProject_createProject;
}

export interface CreateProjectVariables {
  name: string;
  displayName: string;
  site?: string | null;
  description?: string | null;
  photoURL?: string | null;
}
