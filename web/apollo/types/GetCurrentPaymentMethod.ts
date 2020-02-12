/* tslint:disable */
/* eslint-disable */
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: GetCurrentPaymentMethod
// ====================================================

export interface GetCurrentPaymentMethod_getCurrentPaymentMethod_card {
  __typename: "Card";
  brand: string;
  last4: string;
}

export interface GetCurrentPaymentMethod_getCurrentPaymentMethod {
  __typename: "PaymentMethod";
  organizationID: string;
  type: string;
  card: GetCurrentPaymentMethod_getCurrentPaymentMethod_card | null;
}

export interface GetCurrentPaymentMethod {
  getCurrentPaymentMethod: GetCurrentPaymentMethod_getCurrentPaymentMethod | null;
}

export interface GetCurrentPaymentMethodVariables {
  organizationID: ControlUUID;
}
