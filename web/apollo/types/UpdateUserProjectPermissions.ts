/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: UpdateUserProjectPermissions
// ====================================================

export interface UpdateUserProjectPermissions_updateUserProjectPermissions {
  __typename: "PermissionsUsersProjects";
  userID: ControlUUID;
  projectID: ControlUUID;
  view: boolean;
  create: boolean;
  admin: boolean;
}

export interface UpdateUserProjectPermissions {
  updateUserProjectPermissions: UpdateUserProjectPermissions_updateUserProjectPermissions;
}

export interface UpdateUserProjectPermissionsVariables {
  userID: ControlUUID;
  projectID: ControlUUID;
  view?: boolean | null;
  create?: boolean | null;
  admin?: boolean | null;
}
