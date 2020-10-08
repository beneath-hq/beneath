/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: ProjectByOrganizationAndName
// ====================================================

export interface ProjectByOrganizationAndName_projectByOrganizationAndName_organization {
  __typename: "PublicOrganization";
  organizationID: string;
  name: string;
}

export interface ProjectByOrganizationAndName_projectByOrganizationAndName_streams {
  __typename: "Stream";
  streamID: string;
  name: string;
  description: string | null;
}

export interface ProjectByOrganizationAndName_projectByOrganizationAndName_services {
  __typename: "Service";
  serviceID: ControlUUID;
  name: string;
  description: string | null;
  createdOn: ControlTime;
}

export interface ProjectByOrganizationAndName_projectByOrganizationAndName_permissions {
  __typename: "PermissionsUsersProjects";
  view: boolean;
  create: boolean;
  admin: boolean;
}

export interface ProjectByOrganizationAndName_projectByOrganizationAndName {
  __typename: "Project";
  projectID: string;
  name: string;
  displayName: string;
  site: string | null;
  description: string | null;
  photoURL: string | null;
  public: boolean;
  createdOn: ControlTime;
  updatedOn: ControlTime;
  organization: ProjectByOrganizationAndName_projectByOrganizationAndName_organization;
  streams: ProjectByOrganizationAndName_projectByOrganizationAndName_streams[];
  services: ProjectByOrganizationAndName_projectByOrganizationAndName_services[];
  permissions: ProjectByOrganizationAndName_projectByOrganizationAndName_permissions;
}

export interface ProjectByOrganizationAndName {
  projectByOrganizationAndName: ProjectByOrganizationAndName_projectByOrganizationAndName;
}

export interface ProjectByOrganizationAndNameVariables {
  organizationName: string;
  projectName: string;
}
