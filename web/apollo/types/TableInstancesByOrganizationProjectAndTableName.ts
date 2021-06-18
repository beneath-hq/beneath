/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: TableInstancesByOrganizationProjectAndTableName
// ====================================================

export interface TableInstancesByOrganizationProjectAndTableName_tableInstancesByOrganizationProjectAndTableName {
  __typename: "TableInstance";
  tableInstanceID: string;
  tableID: ControlUUID;
  version: number;
  createdOn: ControlTime;
  madePrimaryOn: ControlTime | null;
  madeFinalOn: ControlTime | null;
}

export interface TableInstancesByOrganizationProjectAndTableName {
  tableInstancesByOrganizationProjectAndTableName: TableInstancesByOrganizationProjectAndTableName_tableInstancesByOrganizationProjectAndTableName[];
}

export interface TableInstancesByOrganizationProjectAndTableNameVariables {
  organizationName: string;
  projectName: string;
  tableName: string;
}
