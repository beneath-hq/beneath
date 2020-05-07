/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: OrganizationMembers
// ====================================================

export interface OrganizationMembers_organizationMembers {
  __typename: "OrganizationMember";
  userID: string;
  billingOrganizationID: ControlUUID;
  name: string;
  displayName: string;
  photoURL: string | null;
  view: boolean;
  create: boolean;
  admin: boolean;
  readQuota: number | null;
  writeQuota: number | null;
}

export interface OrganizationMembers {
  organizationMembers: OrganizationMembers_organizationMembers[];
}

export interface OrganizationMembersVariables {
  organizationID: ControlUUID;
}
