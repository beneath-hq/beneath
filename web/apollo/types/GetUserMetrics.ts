/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: GetUserMetrics
// ====================================================

export interface GetUserMetrics_getUserMetrics {
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
}

export interface GetUserMetrics {
  getUserMetrics: GetUserMetrics_getUserMetrics[];
}

export interface GetUserMetricsVariables {
  userID: ControlUUID;
  period: string;
  from: ControlTime;
  until?: ControlTime | null;
}
