/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: GetOrganizationUsage
// ====================================================

export interface GetOrganizationUsage_getOrganizationUsage {
  __typename: "Usage";
  entityID: ControlUUID;
  period: string;
  time: ControlTime;
  readOps: number;
  readBytes: number;
  readRecords: number;
  writeOps: number;
  writeBytes: number;
  writeRecords: number;
  scanOps: number;
  scanBytes: number;
}

export interface GetOrganizationUsage {
  getOrganizationUsage: GetOrganizationUsage_getOrganizationUsage[];
}

export interface GetOrganizationUsageVariables {
  organizationID: ControlUUID;
  period: string;
  from: ControlTime;
  until?: ControlTime | null;
}
