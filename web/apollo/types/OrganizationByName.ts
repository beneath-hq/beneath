/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: OrganizationByName
// ====================================================

export interface OrganizationByName_organizationByName_PublicOrganization_projects {
  __typename: "Project";
  projectID: string;
  name: string;
  displayName: string;
  description: string | null;
  photoURL: string | null;
}

export interface OrganizationByName_organizationByName_PublicOrganization {
  __typename: "PublicOrganization";
  organizationID: string;
  name: string;
  displayName: string;
  description: string | null;
  photoURL: string | null;
  createdOn: ControlTime;
  projects: OrganizationByName_organizationByName_PublicOrganization_projects[];
  personalUserID: ControlUUID | null;
}

export interface OrganizationByName_organizationByName_PrivateOrganization_projects {
  __typename: "Project";
  projectID: string;
  name: string;
  displayName: string;
  description: string | null;
  photoURL: string | null;
}

export interface OrganizationByName_organizationByName_PrivateOrganization_services {
  __typename: "Service";
  serviceID: ControlUUID;
  name: string;
  kind: string;
}

export interface OrganizationByName_organizationByName_PrivateOrganization_personalUser_billingOrganization {
  __typename: "PublicOrganization";
  organizationID: string;
  name: string;
  displayName: string;
}

export interface OrganizationByName_organizationByName_PrivateOrganization_personalUser {
  __typename: "PrivateUser";
  userID: string;
  email: string;
  createdOn: ControlTime;
  updatedOn: ControlTime;
  consentTerms: boolean;
  consentNewsletter: boolean;
  readQuota: number | null;
  writeQuota: number | null;
  billingOrganizationID: ControlUUID;
  billingOrganization: OrganizationByName_organizationByName_PrivateOrganization_personalUser_billingOrganization;
}

export interface OrganizationByName_organizationByName_PrivateOrganization_permissions {
  __typename: "PermissionsUsersOrganizations";
  view: boolean;
  create: boolean;
  admin: boolean;
}

export interface OrganizationByName_organizationByName_PrivateOrganization {
  __typename: "PrivateOrganization";
  organizationID: string;
  name: string;
  displayName: string;
  description: string | null;
  photoURL: string | null;
  createdOn: ControlTime;
  projects: OrganizationByName_organizationByName_PrivateOrganization_projects[];
  personalUserID: ControlUUID | null;
  updatedOn: ControlTime;
  prepaidReadQuota: number | null;
  prepaidWriteQuota: number | null;
  readQuota: number | null;
  writeQuota: number | null;
  readUsage: number;
  writeUsage: number;
  services: OrganizationByName_organizationByName_PrivateOrganization_services[];
  personalUser: OrganizationByName_organizationByName_PrivateOrganization_personalUser | null;
  permissions: OrganizationByName_organizationByName_PrivateOrganization_permissions;
}

export type OrganizationByName_organizationByName = OrganizationByName_organizationByName_PublicOrganization | OrganizationByName_organizationByName_PrivateOrganization;

export interface OrganizationByName {
  organizationByName: OrganizationByName_organizationByName;
}

export interface OrganizationByNameVariables {
  name: string;
}
