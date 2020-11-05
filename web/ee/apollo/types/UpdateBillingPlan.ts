/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: UpdateBillingPlan
// ====================================================

export interface UpdateBillingPlan_updateBillingPlan_billingPlan {
  __typename: "BillingPlan";
  billingPlanID: string;
  default: boolean;
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
  availableInUI: boolean;
}

export interface UpdateBillingPlan_updateBillingPlan {
  __typename: "BillingInfo";
  organizationID: string;
  billingPlan: UpdateBillingPlan_updateBillingPlan_billingPlan;
}

export interface UpdateBillingPlan {
  updateBillingPlan: UpdateBillingPlan_updateBillingPlan;
}

export interface UpdateBillingPlanVariables {
  organizationID: ControlUUID;
  billingPlanID: ControlUUID;
}
