/* tslint:disable */
/* eslint-disable */
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: BillingInfo
// ====================================================

export interface BillingInfo_billingInfo_billingPlan {
  __typename: "BillingPlan";
  billingPlanID: ControlUUID;
  description: string | null;
  currency: string;
  period: string;
  seatPriceCents: number;
  seatReadQuota: number;
  seatWriteQuota: number;
  readOveragePriceCents: number;
  writeOveragePriceCents: number;
  baseReadQuota: number;
  baseWriteQuota: number;
}

export interface BillingInfo_billingInfo {
  __typename: "BillingInfo";
  organizationID: ControlUUID;
  billingPlan: BillingInfo_billingInfo_billingPlan;
  paymentsDriver: string;
}

export interface BillingInfo {
  billingInfo: BillingInfo_billingInfo;
}

export interface BillingInfoVariables {
  organizationID: ControlUUID;
}
