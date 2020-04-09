/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: UpdateBillingInfoBillingPlan
// ====================================================

export interface UpdateBillingInfoBillingPlan_updateBillingInfoBillingPlan_billingPlan {
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

export interface UpdateBillingInfoBillingPlan_updateBillingInfoBillingPlan_billingMethod {
  __typename: "BillingMethod";
  paymentsDriver: string;
}

export interface UpdateBillingInfoBillingPlan_updateBillingInfoBillingPlan {
  __typename: "BillingInfo";
  organizationID: ControlUUID;
  billingPlan: UpdateBillingInfoBillingPlan_updateBillingInfoBillingPlan_billingPlan;
  billingMethod: UpdateBillingInfoBillingPlan_updateBillingInfoBillingPlan_billingMethod;
}

export interface UpdateBillingInfoBillingPlan {
  updateBillingInfoBillingPlan: UpdateBillingInfoBillingPlan_updateBillingInfoBillingPlan;
}

export interface UpdateBillingInfoBillingPlanVariables {
  organizationID: ControlUUID;
  billingPlanID: ControlUUID;
}
