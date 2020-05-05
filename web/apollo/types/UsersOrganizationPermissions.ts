/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: UsersOrganizationPermissions
// ====================================================

export interface UsersOrganizationPermissions_usersOrganizationPermissions {
  __typename: "PermissionsUsersOrganizations";
  userID: ControlUUID;
  organizationID: ControlUUID;
  view: boolean;
  admin: boolean;
}

export interface UsersOrganizationPermissions {
  usersOrganizationPermissions: UsersOrganizationPermissions_usersOrganizationPermissions[];
}

export interface UsersOrganizationPermissionsVariables {
  organizationID: ControlUUID;
}
