/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: Me
// ====================================================

export interface Me_me_projects {
  __typename: "Project";
  projectID: string;
  name: string;
  displayName: string;
  description: string | null;
  photoURL: string | null;
}

export interface Me_me_personalUser_billingOrganization {
  __typename: "PublicOrganization";
  organizationID: string;
  name: string;
  displayName: string;
}

export interface Me_me_personalUser {
  __typename: "PrivateUser";
  userID: string;
  email: string;
  createdOn: ControlTime;
  updatedOn: ControlTime;
  consentTerms: boolean;
  consentNewsletter: boolean;
  readQuota: number | null;
  writeQuota: number | null;
  scanQuota: number | null;
  billingOrganizationID: ControlUUID;
  billingOrganization: Me_me_personalUser_billingOrganization;
}

export interface Me_me_permissions {
  __typename: "PermissionsUsersOrganizations";
  view: boolean;
  create: boolean;
  admin: boolean;
}

export interface Me_me {
  __typename: "PrivateOrganization";
  organizationID: string;
  name: string;
  displayName: string;
  description: string | null;
  photoURL: string | null;
  createdOn: ControlTime;
  projects: Me_me_projects[];
  personalUserID: ControlUUID | null;
  updatedOn: ControlTime;
  prepaidReadQuota: number | null;
  prepaidWriteQuota: number | null;
  prepaidScanQuota: number | null;
  readQuota: number | null;
  writeQuota: number | null;
  scanQuota: number | null;
  readUsage: number;
  writeUsage: number;
  personalUser: Me_me_personalUser | null;
  permissions: Me_me_permissions;
}

export interface Me {
  me: Me_me | null;
}
