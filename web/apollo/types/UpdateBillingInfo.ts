/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: UpdateBillingInfo
// ====================================================

export interface UpdateBillingInfo_updateBillingInfo_billingPlan {
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

export interface UpdateBillingInfo_updateBillingInfo_billingMethod {
  __typename: "BillingMethod";
  paymentsDriver: string;
  driverPayload: string;
}

export interface UpdateBillingInfo_updateBillingInfo {
  __typename: "BillingInfo";
  organizationID: ControlUUID;
  billingPlan: UpdateBillingInfo_updateBillingInfo_billingPlan;
  billingMethod: UpdateBillingInfo_updateBillingInfo_billingMethod;
}

export interface UpdateBillingInfo {
  updateBillingInfo: UpdateBillingInfo_updateBillingInfo;
}

export interface UpdateBillingInfoVariables {
  organizationID: ControlUUID;
  billingMethodID: ControlUUID;
  billingPlanID: ControlUUID;
}
