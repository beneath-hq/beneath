import gql from "graphql-tag";

export const QUERY_SERVICE = gql`
  query ServiceByOrganizationProjectAndName($organizationName: String!, $projectName: String!, $serviceName: String!) {
    serviceByOrganizationProjectAndName(
      organizationName: $organizationName
      projectName: $projectName
      serviceName: $serviceName
    ) {
      serviceID
      name
      description
      sourceURL
      project {
        projectID
        public
        permissions {
          view
          create
          admin
        }
      }
      quotaStartTime
      quotaEndTime
      readQuota
      writeQuota
      scanQuota
    }
  }
`;

export const QUERY_STREAM_PERMISSIONS_FOR_SERVICE = gql`
  query TablePermissionsForService($serviceID: UUID!) {
    tablePermissionsForService(serviceID: $serviceID) {
      serviceID
      tableID
      read
      write
      table {
        tableID
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
  mutation UpdateServiceTablePermissions($serviceID: UUID!, $tableID: UUID!, $read: Boolean, $write: Boolean) {
    updateServiceTablePermissions(serviceID: $serviceID, tableID: $tableID, read: $read, write: $write) {
      serviceID
      tableID
      read
      write
      table {
        tableID
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
