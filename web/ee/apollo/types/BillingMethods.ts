/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: BillingMethods
// ====================================================

export interface BillingMethods_billingMethods {
  __typename: "BillingMethod";
  billingMethodID: string;
  organizationID: ControlUUID;
  paymentsDriver: string;
  driverPayload: string;
}

export interface BillingMethods {
  billingMethods: BillingMethods_billingMethods[];
}

export interface BillingMethodsVariables {
  organizationID: ControlUUID;
}
