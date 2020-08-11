/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { EntityKind } from "./globalTypes";

// ====================================================
// GraphQL query operation: GetMetrics
// ====================================================

export interface GetMetrics_getMetrics {
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
  scanOps: number;
  scanBytes: number;
}

export interface GetMetrics {
  getMetrics: GetMetrics_getMetrics[];
}

export interface GetMetricsVariables {
  entityKind: EntityKind;
  entityID: ControlUUID;
  period: string;
  from: ControlTime;
  until?: ControlTime | null;
}
