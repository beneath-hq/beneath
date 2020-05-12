/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: ProjectsForUser
// ====================================================

export interface ProjectsForUser_projectsForUser_organization {
  __typename: "PublicOrganization";
  organizationID: string;
  name: string;
}

export interface ProjectsForUser_projectsForUser {
  __typename: "Project";
  projectID: string;
  name: string;
  displayName: string;
  description: string | null;
  photoURL: string | null;
  createdOn: ControlTime;
  updatedOn: ControlTime;
  organization: ProjectsForUser_projectsForUser_organization;
}

export interface ProjectsForUser {
  projectsForUser: ProjectsForUser_projectsForUser[] | null;
}

export interface ProjectsForUserVariables {
  userID: ControlUUID;
}
