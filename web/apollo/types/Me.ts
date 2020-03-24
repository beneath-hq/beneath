/* tslint:disable */
/* eslint-disable */
// @generated
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
  username: string;
  name: string;
  bio: string | null;
  photoURL: string | null;
  createdOn: ControlTime;
  projects: Me_me_user_projects[];
}

export interface Me_me_organization {
  __typename: "Organization";
  organizationID: string;
  name: string;
  personal: boolean;
}

export interface Me_me {
  __typename: "Me";
  userID: string;
  email: string;
  readUsage: number;
  readQuota: number;
  writeUsage: number;
  writeQuota: number;
  updatedOn: ControlTime;
  user: Me_me_user;
  organization: Me_me_organization;
}

export interface Me {
  me: Me_me | null;
}
