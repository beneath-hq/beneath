/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: OrganizationByName
// ====================================================

export interface OrganizationByName_organizationByName_users {
  __typename: "User";
  userID: string;
  name: string;
  username: string;
  photoURL: string | null;
  readQuota: number;
  writeQuota: number;
}

export interface OrganizationByName_organizationByName_services {
  __typename: "Service";
  serviceID: ControlUUID;
  name: string;
  kind: string;
}

export interface OrganizationByName_organizationByName {
  __typename: "Organization";
  organizationID: string;
  name: string;
  createdOn: ControlTime;
  updatedOn: ControlTime;
  users: OrganizationByName_organizationByName_users[];
  services: OrganizationByName_organizationByName_services[];
}

export interface OrganizationByName {
  organizationByName: OrganizationByName_organizationByName | null;
}

export interface OrganizationByNameVariables {
  name: string;
}
