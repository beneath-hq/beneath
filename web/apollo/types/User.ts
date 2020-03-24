/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: User
// ====================================================

export interface User_user_projects {
  __typename: "Project";
  projectID: string;
  name: string;
  displayName: string;
  description: string | null;
  photoURL: string | null;
}

export interface User_user {
  __typename: "User";
  userID: string;
  username: string;
  name: string;
  bio: string | null;
  photoURL: string | null;
  createdOn: ControlTime;
  projects: User_user_projects[];
}

export interface User {
  user: User_user | null;
}

export interface UserVariables {
  userID: ControlUUID;
}
