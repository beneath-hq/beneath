/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: RegisterUserConsent
// ====================================================

export interface RegisterUserConsent_registerUserConsent {
  __typename: "PrivateUser";
  userID: string;
  updatedOn: ControlTime;
  consentTerms: boolean;
  consentNewsletter: boolean;
}

export interface RegisterUserConsent {
  registerUserConsent: RegisterUserConsent_registerUserConsent;
}

export interface RegisterUserConsentVariables {
  userID: ControlUUID;
  terms?: boolean | null;
  newsletter?: boolean | null;
}
