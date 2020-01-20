/* tslint:disable */
/* eslint-disable */
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: AddUserToOrganization
// ====================================================

export interface AddUserToOrganization_addUserToOrganization {
  __typename: "User";
  userID: string;
}

export interface AddUserToOrganization {
  addUserToOrganization: AddUserToOrganization_addUserToOrganization | null;
}

export interface AddUserToOrganizationVariables {
  username: string;
  organizationID: ControlUUID;
  view: boolean;
  admin: boolean;
}
