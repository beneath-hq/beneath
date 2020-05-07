/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: BillingPlans
// ====================================================

export interface BillingPlans_billingPlans {
  __typename: "BillingPlan";
  billingPlanID: ControlUUID;
  default: boolean;
  description: string | null;
  availableInUI: boolean;
}

export interface BillingPlans {
  billingPlans: BillingPlans_billingPlans[];
}
