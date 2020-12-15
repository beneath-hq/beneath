/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { GetUsageInput, UsageLabel } from "./globalTypes";

// ====================================================
// GraphQL query operation: GetUsage
// ====================================================

export interface GetUsage_getUsage {
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

export interface GetUsage {
  getUsage: GetUsage_getUsage[];
}

export interface GetUsageVariables {
  input: GetUsageInput;
}
