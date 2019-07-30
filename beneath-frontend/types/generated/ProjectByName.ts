/* tslint:disable */
/* eslint-disable */
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: ProjectByName
// ====================================================

export interface ProjectByName_projectByName_users {
  __typename: "User";
  userID: string;
  name: string;
  username: string | null;
  photoURL: string | null;
}

export interface ProjectByName_projectByName_streams {
  __typename: "Stream";
  streamID: string;
  name: string;
  description: string | null;
  external: boolean;
}

export interface ProjectByName_projectByName {
  __typename: "Project";
  projectID: string;
  name: string;
  displayName: string;
  site: string | null;
  description: string | null;
  photoURL: string | null;
  createdOn: ControlTime;
  updatedOn: ControlTime;
  users: ProjectByName_projectByName_users[];
  streams: ProjectByName_projectByName_streams[];
}

export interface ProjectByName {
  projectByName: ProjectByName_projectByName | null;
}

export interface ProjectByNameVariables {
  name: string;
}
