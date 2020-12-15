/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { GetEntityUsageInput, UsageLabel } from "./globalTypes";

// ====================================================
// GraphQL query operation: GetUserUsage
// ====================================================

export interface GetUserUsage_getUserUsage {
  __typename: "Usage";
  entityID: ControlUUID;
  label: UsageLabel;
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
  input: GetEntityUsageInput;
}
