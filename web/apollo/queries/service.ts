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

export const QUERY_STREAM_PERMISSIONS_FOR_SERVICE = gql`
  query StreamPermissionsForService($serviceID: UUID!) {
    streamPermissionsForService(serviceID: $serviceID) {
      serviceID
    	streamID
      read
      write
      stream {
        streamID
        name
        project {
          projectID
          name
          organization {
            organizationID
            name
          }
        }
      }
    }
  }
`;

export const CREATE_SERVICE = gql`
  mutation CreateService($input: CreateServiceInput!) {
    createService(input: $input) {
      serviceID
      name
      description
      sourceURL
      readQuota
      writeQuota
      scanQuota
      project {
        projectID
        name
        organization {
          organizationID
          name
        }
      }
    }
  }
`;

export const UPDATE_SERVICE = gql`
  mutation UpdateService($input: UpdateServiceInput!) {
    updateService(input: $input) {
      serviceID
      name
      description
      sourceURL
      readQuota
      writeQuota
      scanQuota
      project {
        projectID
        name
        organization {
          organizationID
          name
        }
      }
    }
  }
`;

export const UPDATE_SERVICE_STREAM_PERMISSIONS = gql`
  mutation UpdateServiceStreamPermissions($serviceID: UUID!, $streamID: UUID!, $read: Boolean, $write: Boolean) {
    updateServiceStreamPermissions(serviceID: $serviceID, streamID: $streamID, read: $read, write: $write) {
      serviceID
      streamID
      read
      write
      stream {
        streamID
        name
        project {
          projectID
          name
          organization {
            organizationID
            name
          }
        }
      }
    }
  }
`;