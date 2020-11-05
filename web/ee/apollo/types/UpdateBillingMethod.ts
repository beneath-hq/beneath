/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: UpdateBillingMethod
// ====================================================

export interface UpdateBillingMethod_updateBillingMethod_billingMethod {
  __typename: "BillingMethod";
  billingMethodID: string;
  paymentsDriver: string;
  driverPayload: string;
}

export interface UpdateBillingMethod_updateBillingMethod {
  __typename: "BillingInfo";
  organizationID: string;
  billingMethod: UpdateBillingMethod_updateBillingMethod_billingMethod | null;
}

export interface UpdateBillingMethod {
  updateBillingMethod: UpdateBillingMethod_updateBillingMethod;
}

export interface UpdateBillingMethodVariables {
  organizationID: ControlUUID;
  billingMethodID?: ControlUUID | null;
}
