/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: ProjectByOrganizationAndName
// ====================================================

export interface ProjectByOrganizationAndName_projectByOrganizationAndName_organization {
  __typename: "Organization";
  name: string;
}

export interface ProjectByOrganizationAndName_projectByOrganizationAndName_users {
  __typename: "User";
  userID: string;
  name: string;
  username: string;
  photoURL: string | null;
}

export interface ProjectByOrganizationAndName_projectByOrganizationAndName_streams {
  __typename: "Stream";
  streamID: string;
  name: string;
  description: string | null;
  external: boolean;
}

export interface ProjectByOrganizationAndName_projectByOrganizationAndName {
  __typename: "Project";
  projectID: string;
  name: string;
  displayName: string;
  site: string | null;
  description: string | null;
  photoURL: string | null;
  createdOn: ControlTime;
  updatedOn: ControlTime;
  organization: ProjectByOrganizationAndName_projectByOrganizationAndName_organization;
  users: ProjectByOrganizationAndName_projectByOrganizationAndName_users[];
  streams: ProjectByOrganizationAndName_projectByOrganizationAndName_streams[];
}

export interface ProjectByOrganizationAndName {
  projectByOrganizationAndName: ProjectByOrganizationAndName_projectByOrganizationAndName | null;
}

export interface ProjectByOrganizationAndNameVariables {
  organizationName: string;
  projectName: string;
}
