/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: UpdateOrganization
// ====================================================

export interface UpdateOrganization_updateOrganization_personalUser {
  __typename: "PrivateUser";
  userID: string;
  email: string;
  createdOn: ControlTime;
  updatedOn: ControlTime;
  readQuota: number | null;
  writeQuota: number | null;
  scanQuota: number | null;
  billingOrganizationID: ControlUUID;
}

export interface UpdateOrganization_updateOrganization {
  __typename: "PrivateOrganization";
  organizationID: string;
  name: string;
  displayName: string;
  description: string | null;
  photoURL: string | null;
  createdOn: ControlTime;
  updatedOn: ControlTime;
  personalUserID: ControlUUID | null;
  personalUser: UpdateOrganization_updateOrganization_personalUser | null;
}

export interface UpdateOrganization {
  updateOrganization: UpdateOrganization_updateOrganization;
}

export interface UpdateOrganizationVariables {
  organizationID: ControlUUID;
  name?: string | null;
  displayName?: string | null;
  description?: string | null;
  photoURL?: string | null;
}
