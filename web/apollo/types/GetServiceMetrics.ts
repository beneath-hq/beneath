/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: GetServiceMetrics
// ====================================================

export interface GetServiceMetrics_getServiceMetrics {
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

export interface GetServiceMetrics {
  getServiceMetrics: GetServiceMetrics_getServiceMetrics[];
}

export interface GetServiceMetricsVariables {
  serviceID: ControlUUID;
  period: string;
  from: ControlTime;
  until?: ControlTime | null;
}
