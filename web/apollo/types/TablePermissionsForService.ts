/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: TablePermissionsForService
// ====================================================

export interface TablePermissionsForService_tablePermissionsForService_table_project_organization {
  __typename: "PublicOrganization";
  organizationID: string;
  name: string;
}

export interface TablePermissionsForService_tablePermissionsForService_table_project {
  __typename: "Project";
  projectID: string;
  name: string;
  organization: TablePermissionsForService_tablePermissionsForService_table_project_organization;
}

export interface TablePermissionsForService_tablePermissionsForService_table {
  __typename: "Table";
  tableID: string;
  name: string;
  project: TablePermissionsForService_tablePermissionsForService_table_project;
}

export interface TablePermissionsForService_tablePermissionsForService {
  __typename: "PermissionsServicesTables";
  serviceID: ControlUUID;
  tableID: ControlUUID;
  read: boolean;
  write: boolean;
  table: TablePermissionsForService_tablePermissionsForService_table | null;
}

export interface TablePermissionsForService {
  tablePermissionsForService: TablePermissionsForService_tablePermissionsForService[];
}

export interface TablePermissionsForServiceVariables {
  serviceID: ControlUUID;
}
