/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: BillingInfo
// ====================================================

export interface BillingInfo_billingInfo_billingPlan {
  __typename: "BillingPlan";
  default: boolean;
  billingPlanID: ControlUUID;
  description: string | null;
  currency: string;
  period: string;
  seatPriceCents: number;
  seatReadQuota: number;
  seatWriteQuota: number;
  seatScanQuota: number;
  readOveragePriceCents: number;
  writeOveragePriceCents: number;
  scanOveragePriceCents: number;
  baseReadQuota: number;
  baseWriteQuota: number;
  baseScanQuota: number;
  availableInUI: boolean;
}

export interface BillingInfo_billingInfo_billingMethod {
  __typename: "BillingMethod";
  billingMethodID: ControlUUID;
  paymentsDriver: string;
  driverPayload: string;
}

export interface BillingInfo_billingInfo {
  __typename: "BillingInfo";
  organizationID: ControlUUID;
  billingPlan: BillingInfo_billingInfo_billingPlan;
  billingMethod: BillingInfo_billingInfo_billingMethod | null;
  country: string;
  region: string | null;
  companyName: string | null;
  taxNumber: string | null;
}

export interface BillingInfo {
  billingInfo: BillingInfo_billingInfo;
}

export interface BillingInfoVariables {
  organizationID: ControlUUID;
}
