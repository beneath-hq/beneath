import gql from "graphql-tag";

export const QUERY_SERVICE = gql`
  query ServiceByOrganizationProjectAndName($organizationName: String!, $projectName: String!, $serviceName: String!) {
    serviceByOrganizationProjectAndName(
      organizationName: $organizationName,
      projectName: $projectName,
      serviceName: $serviceName,
    ) {
      serviceID
      name
      description
      sourceURL
      quotaStartTime
      quotaEndTime
      readQuota
      writeQuota
      scanQuota
    }
  }
`;
