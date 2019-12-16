/* tslint:disable */
/* eslint-disable */
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: BillingInfo
// ====================================================

export interface BillingInfo_billingInfo {
  __typename: "BillingInfo";
  organizationID: ControlUUID;
  billingPlanID: ControlUUID;
  paymentsDriver: string;
}

export interface BillingInfo {
  billingInfo: BillingInfo_billingInfo;
}

export interface BillingInfoVariables {
  organizationID: ControlUUID;
}
