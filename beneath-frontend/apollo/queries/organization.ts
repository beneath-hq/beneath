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

export const CREATE_STRIPE_SETUP_INTENT = gql`
  mutation CreateStripeSetupIntent($organizationID: UUID!, $billingPlanID: UUID!) {
    createStripeSetupIntent(organizationID: $organizationID, billingPlanID: $billingPlanID)
  }
`;