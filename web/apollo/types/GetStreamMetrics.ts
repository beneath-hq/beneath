/* tslint:disable */
/* eslint-disable */
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: GetStreamMetrics
// ====================================================

export interface GetStreamMetrics_getStreamMetrics {
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

export interface GetStreamMetrics {
  getStreamMetrics: GetStreamMetrics_getStreamMetrics[];
}

export interface GetStreamMetricsVariables {
  streamID: ControlUUID;
  period: string;
  from: ControlTime;
  until?: ControlTime | null;
}
