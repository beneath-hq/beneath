import gql from "graphql-tag";

export const QUERY_SERVICE = gql`
  query ServiceByNameAndOrganization($serviceName: String!, $organizationName: String!) {
    serviceByNameAndOrganization(name: $serviceName, organizationName: $organizationName) {
      serviceID
      name
      kind
      readQuota
      writeQuota
    }
  }
`;
