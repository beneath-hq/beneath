/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { GetEntityUsageInput, UsageLabel } from "./globalTypes";

// ====================================================
// GraphQL query operation: GetStreamUsage
// ====================================================

export interface GetStreamUsage_getStreamUsage {
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
}

export interface GetStreamUsage {
  getStreamUsage: GetStreamUsage_getStreamUsage[];
}

export interface GetStreamUsageVariables {
  input: GetEntityUsageInput;
}
