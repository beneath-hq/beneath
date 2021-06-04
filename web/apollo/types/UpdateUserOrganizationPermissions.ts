/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: UpdateUserOrganizationPermissions
// ====================================================

export interface UpdateUserOrganizationPermissions_updateUserOrganizationPermissions {
  __typename: "PermissionsUsersOrganizations";
  userID: ControlUUID;
  organizationID: ControlUUID;
  view: boolean;
  create: boolean;
  admin: boolean;
}

export interface UpdateUserOrganizationPermissions {
  updateUserOrganizationPermissions: UpdateUserOrganizationPermissions_updateUserOrganizationPermissions;
}

export interface UpdateUserOrganizationPermissionsVariables {
  userID: ControlUUID;
  organizationID: ControlUUID;
  view?: boolean | null;
  create?: boolean | null;
  admin?: boolean | null;
}
