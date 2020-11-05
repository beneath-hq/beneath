/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: GetStreamUsage
// ====================================================

export interface GetStreamUsage_getStreamUsage {
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
}

export interface GetStreamUsage {
  getStreamUsage: GetStreamUsage_getStreamUsage[];
}

export interface GetStreamUsageVariables {
  streamID: ControlUUID;
  period: string;
  from: ControlTime;
  until?: ControlTime | null;
}
