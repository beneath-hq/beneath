import gql from "graphql-tag";

export const QUERY_BILLING_METHODS = gql`
  query BillingMethods($organizationID: UUID!){
    billingMethods(organizationID: $organizationID) {
      billingMethodID
      organizationID
      paymentsDriver
      driverPayload
    }
  }
`;
