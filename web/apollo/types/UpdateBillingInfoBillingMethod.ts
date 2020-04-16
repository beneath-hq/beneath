/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: UpdateBillingInfoBillingMethod
// ====================================================

export interface UpdateBillingInfoBillingMethod_updateBillingInfoBillingMethod_billingPlan {
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

export interface UpdateBillingInfoBillingMethod_updateBillingInfoBillingMethod_billingMethod {
  __typename: "BillingMethod";
  paymentsDriver: string;
  driverPayload: string;
}

export interface UpdateBillingInfoBillingMethod_updateBillingInfoBillingMethod {
  __typename: "BillingInfo";
  organizationID: ControlUUID;
  billingPlan: UpdateBillingInfoBillingMethod_updateBillingInfoBillingMethod_billingPlan;
  billingMethod: UpdateBillingInfoBillingMethod_updateBillingInfoBillingMethod_billingMethod;
}

export interface UpdateBillingInfoBillingMethod {
  updateBillingInfoBillingMethod: UpdateBillingInfoBillingMethod_updateBillingInfoBillingMethod;
}

export interface UpdateBillingInfoBillingMethodVariables {
  organizationID: ControlUUID;
  billingMethodID: ControlUUID;
}
