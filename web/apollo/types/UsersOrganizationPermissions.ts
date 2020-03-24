/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: UsersOrganizationPermissions
// ====================================================

export interface UsersOrganizationPermissions_usersOrganizationPermissions_user {
  __typename: "User";
  userID: string;
}

export interface UsersOrganizationPermissions_usersOrganizationPermissions_organization {
  __typename: "Organization";
  organizationID: string;
}

export interface UsersOrganizationPermissions_usersOrganizationPermissions {
  __typename: "PermissionsUsersOrganizations";
  user: UsersOrganizationPermissions_usersOrganizationPermissions_user;
  organization: UsersOrganizationPermissions_usersOrganizationPermissions_organization;
  view: boolean;
  admin: boolean;
}

export interface UsersOrganizationPermissions {
  usersOrganizationPermissions: UsersOrganizationPermissions_usersOrganizationPermissions[];
}

export interface UsersOrganizationPermissionsVariables {
  organizationID: ControlUUID;
}
