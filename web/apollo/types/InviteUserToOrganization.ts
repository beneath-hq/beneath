/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: InviteUserToOrganization
// ====================================================

export interface InviteUserToOrganization {
  inviteUserToOrganization: boolean;
}

export interface InviteUserToOrganizationVariables {
  userID: ControlUUID;
  organizationID: ControlUUID;
  view: boolean;
  create: boolean;
  admin: boolean;
}
