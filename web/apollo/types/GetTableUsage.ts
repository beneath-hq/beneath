/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { GetEntityUsageInput, UsageLabel } from "./globalTypes";

// ====================================================
// GraphQL query operation: GetTableUsage
// ====================================================

export interface GetTableUsage_getTableUsage {
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

export interface GetTableUsage {
  getTableUsage: GetTableUsage_getTableUsage[];
}

export interface GetTableUsageVariables {
  input: GetEntityUsageInput;
}
