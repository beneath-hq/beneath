/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: ServiceByOrganizationProjectAndName
// ====================================================

export interface ServiceByOrganizationProjectAndName_serviceByOrganizationProjectAndName_project_permissions {
  __typename: "PermissionsUsersProjects";
  view: boolean;
  create: boolean;
  admin: boolean;
}

export interface ServiceByOrganizationProjectAndName_serviceByOrganizationProjectAndName_project {
  __typename: "Project";
  projectID: string;
  public: boolean;
  permissions: ServiceByOrganizationProjectAndName_serviceByOrganizationProjectAndName_project_permissions;
}

export interface ServiceByOrganizationProjectAndName_serviceByOrganizationProjectAndName {
  __typename: "Service";
  serviceID: string;
  name: string;
  description: string | null;
  sourceURL: string | null;
  project: ServiceByOrganizationProjectAndName_serviceByOrganizationProjectAndName_project;
  quotaStartTime: ControlTime;
  quotaEndTime: ControlTime;
  readQuota: number | null;
  writeQuota: number | null;
  scanQuota: number | null;
}

export interface ServiceByOrganizationProjectAndName {
  serviceByOrganizationProjectAndName: ServiceByOrganizationProjectAndName_serviceByOrganizationProjectAndName;
}

export interface ServiceByOrganizationProjectAndNameVariables {
  organizationName: string;
  projectName: string;
  serviceName: string;
}
