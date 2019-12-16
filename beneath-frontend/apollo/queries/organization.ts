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
