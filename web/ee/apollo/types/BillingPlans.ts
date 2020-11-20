/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: BillingPlans
// ====================================================

export interface BillingPlans_billingPlans {
  __typename: "BillingPlan";
  billingPlanID: string;
  default: boolean;
  name: string;
  description: string | null;
  currency: string;
  period: string;
  basePriceCents: number;
  seatPriceCents: number;
  baseReadQuota: number;
  baseWriteQuota: number;
  baseScanQuota: number;
  seatReadQuota: number;
  seatWriteQuota: number;
  seatScanQuota: number;
  readQuota: number;
  writeQuota: number;
  scanQuota: number;
  readOveragePriceCents: number;
  writeOveragePriceCents: number;
  scanOveragePriceCents: number;
  UIRank: number | null;
}

export interface BillingPlans {
  billingPlans: BillingPlans_billingPlans[];
}
