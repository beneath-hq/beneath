/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: UpdateServiceTablePermissions
// ====================================================

export interface UpdateServiceTablePermissions_updateServiceTablePermissions_table_project_organization {
  __typename: "PublicOrganization";
  organizationID: string;
  name: string;
}

export interface UpdateServiceTablePermissions_updateServiceTablePermissions_table_project {
  __typename: "Project";
  projectID: string;
  name: string;
  organization: UpdateServiceTablePermissions_updateServiceTablePermissions_table_project_organization;
}

export interface UpdateServiceTablePermissions_updateServiceTablePermissions_table {
  __typename: "Table";
  tableID: string;
  name: string;
  project: UpdateServiceTablePermissions_updateServiceTablePermissions_table_project;
}

export interface UpdateServiceTablePermissions_updateServiceTablePermissions {
  __typename: "PermissionsServicesTables";
  serviceID: ControlUUID;
  tableID: ControlUUID;
  read: boolean;
  write: boolean;
  table: UpdateServiceTablePermissions_updateServiceTablePermissions_table | null;
}

export interface UpdateServiceTablePermissions {
  updateServiceTablePermissions: UpdateServiceTablePermissions_updateServiceTablePermissions;
}

export interface UpdateServiceTablePermissionsVariables {
  serviceID: ControlUUID;
  tableID: ControlUUID;
  read?: boolean | null;
  write?: boolean | null;
}
