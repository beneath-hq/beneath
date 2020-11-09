/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: GetUserUsage
// ====================================================

export interface GetUserUsage_getUserUsage {
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

export interface GetUserUsage {
  getUserUsage: GetUserUsage_getUserUsage[];
}

export interface GetUserUsageVariables {
  userID: ControlUUID;
  period: string;
  from: ControlTime;
  until?: ControlTime | null;
}