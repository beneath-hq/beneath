/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: GetOrganizationMetrics
// ====================================================

export interface GetOrganizationMetrics_getOrganizationMetrics {
  __typename: "Metrics";
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

export interface GetOrganizationMetrics {
  getOrganizationMetrics: GetOrganizationMetrics_getOrganizationMetrics[];
}

export interface GetOrganizationMetricsVariables {
  organizationID: ControlUUID;
  period: string;
  from: ControlTime;
  until?: ControlTime | null;
}
