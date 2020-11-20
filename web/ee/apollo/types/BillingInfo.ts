/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: BillingInfo
// ====================================================

export interface BillingInfo_billingInfo_billingPlan {
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

export interface BillingInfo_billingInfo_billingMethod {
  __typename: "BillingMethod";
  billingMethodID: string;
  paymentsDriver: string;
  driverPayload: string;
}

export interface BillingInfo_billingInfo {
  __typename: "BillingInfo";
  organizationID: string;
  billingPlan: BillingInfo_billingInfo_billingPlan;
  billingMethod: BillingInfo_billingInfo_billingMethod | null;
  country: string;
  region: string | null;
  companyName: string | null;
  taxNumber: string | null;
  nextBillingTime: ControlTime;
  lastInvoiceTime: ControlTime;
}

export interface BillingInfo {
  billingInfo: BillingInfo_billingInfo;
}

export interface BillingInfoVariables {
  organizationID: ControlUUID;
}
