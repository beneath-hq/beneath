/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { EntityKind } from "./globalTypes";

// ====================================================
// GraphQL query operation: GetUsage
// ====================================================

export interface GetUsage_getUsage {
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

export interface GetUsage {
  getUsage: GetUsage_getUsage[];
}

export interface GetUsageVariables {
  entityKind: EntityKind;
  entityID: ControlUUID;
  period: string;
  from: ControlTime;
  until?: ControlTime | null;
}
