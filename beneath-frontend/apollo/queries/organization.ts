import gql from "graphql-tag";

export const QUERY_ORGANIZATION = gql`
  query OrganizationByName($name: String!) {
    organizationByName(name: $name) {
      organizationID
      name
      createdOn
      updatedOn
      users {
        userID
        name
        username
        photoURL
      }
      services {
        serviceID
        name
        kind
      }
    }
  }
`;

export const QUERY_CURRENT_PAYMENT_METHOD = gql`
  query GetCurrentPaymentMethod($organizationID: UUID!){
    getCurrentPaymentMethod(organizationID: $organizationID) {
      organizationID
      type
      card {
        brand
        last4
      }    
    }
  }
`;

export const CREATE_STRIPE_SETUP_INTENT = gql`
  mutation CreateStripeSetupIntent($organizationID: UUID!, $billingPlanID: UUID!) {
    createStripeSetupIntent(organizationID: $organizationID, billingPlanID: $billingPlanID)
  }
`;