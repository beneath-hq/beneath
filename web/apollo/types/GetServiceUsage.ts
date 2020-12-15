/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { GetEntityUsageInput, UsageLabel } from "./globalTypes";

// ====================================================
// GraphQL query operation: GetServiceUsage
// ====================================================

export interface GetServiceUsage_getServiceUsage {
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

export interface GetServiceUsage {
  getServiceUsage: GetServiceUsage_getServiceUsage[];
}

export interface GetServiceUsageVariables {
  input: GetEntityUsageInput;
}
