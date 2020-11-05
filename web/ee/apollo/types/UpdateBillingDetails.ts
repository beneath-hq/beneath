/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: UpdateBillingDetails
// ====================================================

export interface UpdateBillingDetails_updateBillingDetails {
  __typename: "BillingInfo";
  organizationID: string;
  country: string;
  region: string | null;
  companyName: string | null;
  taxNumber: string | null;
}

export interface UpdateBillingDetails {
  updateBillingDetails: UpdateBillingDetails_updateBillingDetails;
}

export interface UpdateBillingDetailsVariables {
  organizationID: ControlUUID;
  country?: string | null;
  region?: string | null;
  companyName?: string | null;
  taxNumber?: string | null;
}
