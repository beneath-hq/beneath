import gql from "graphql-tag";

export const QUERY_BILLING_PLANS = gql`
  query BillingPlans{
  billingPlans {
    billingPlanID
    default
    description
  }
}
`;