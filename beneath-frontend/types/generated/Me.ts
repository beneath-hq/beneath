/* tslint:disable */
/* eslint-disable */
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: Me
// ====================================================

export interface Me_me_user_projects {
  __typename: "Project";
  projectID: string;
  name: string;
  displayName: string;
  description: string | null;
  photoURL: string | null;
}

export interface Me_me_user {
  __typename: "User";
  userID: string;
  username: string | null;
  name: string;
  bio: string | null;
  photoURL: string | null;
  createdOn: ControlTime;
  projects: Me_me_user_projects[];
}

export interface Me_me {
  __typename: "Me";
  userID: string;
  email: string;
  updatedOn: ControlTime;
  user: Me_me_user;
}

export interface Me {
  me: Me_me | null;
}
