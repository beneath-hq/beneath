import gql from "graphql-tag";

export const QUERY_BILLING_INFO = gql`
  query BillingInfo($organizationID: UUID!){
  billingInfo(organizationID: $organizationID) {
    organizationID
    billingPlanID
    paymentsDriver
  }
}
`;