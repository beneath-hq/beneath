/* tslint:disable */
/* eslint-disable */
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: InviteUserToOrganization
// ====================================================

export interface InviteUserToOrganization_inviteUserToOrganization {
  __typename: "User";
  userID: string;
}

export interface InviteUserToOrganization {
  inviteUserToOrganization: InviteUserToOrganization_inviteUserToOrganization | null;
}

export interface InviteUserToOrganizationVariables {
  username: string;
  organizationID: ControlUUID;
  view: boolean;
  admin: boolean;
}
