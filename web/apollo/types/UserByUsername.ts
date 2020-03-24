/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: UserByUsername
// ====================================================

export interface UserByUsername_userByUsername_projects {
  __typename: "Project";
  projectID: string;
  name: string;
  displayName: string;
  description: string | null;
  photoURL: string | null;
}

export interface UserByUsername_userByUsername {
  __typename: "User";
  userID: string;
  username: string;
  name: string;
  bio: string | null;
  photoURL: string | null;
  createdOn: ControlTime;
  projects: UserByUsername_userByUsername_projects[];
}

export interface UserByUsername {
  userByUsername: UserByUsername_userByUsername | null;
}

export interface UserByUsernameVariables {
  username: string;
}
