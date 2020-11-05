import gql from "graphql-tag";

export const GET_USAGE = gql`
  query GetUsage($entityKind: EntityKind!, $entityID: UUID!, $period: String!, $from: Time!, $until: Time) {
    getUsage(entityKind: $entityKind, entityID: $entityID, period: $period, from: $from, until: $until) {
      entityID
      period
      time
      readOps
      readBytes
      readRecords
      writeOps
      writeBytes
      writeRecords
      scanOps
      scanBytes
    }
  }
`;

export const GET_ORGANIZATION_USAGE = gql`
  query GetOrganizationUsage($organizationID: UUID!, $period: String!, $from: Time!, $until: Time) {
    getOrganizationUsage(organizationID: $organizationID, period: $period, from: $from, until: $until) {
      entityID
      period
      time
      readOps
      readBytes
      readRecords
      writeOps
      writeBytes
      writeRecords
      scanOps
      scanBytes
    }
  }
`;

export const GET_SERVICE_USAGE = gql`
  query GetServiceUsage($serviceID: UUID!, $period: String!, $from: Time!, $until: Time) {
    getServiceUsage(serviceID: $serviceID, period: $period, from: $from, until: $until) {
      entityID
      period
      time
      readOps
      readBytes
      readRecords
      writeOps
      writeBytes
      writeRecords
      scanOps
      scanBytes
    }
  }
`;

export const GET_STREAM_USAGE = gql`
  query GetStreamUsage($streamID: UUID!, $period: String!, $from: Time!, $until: Time) {
    getStreamUsage(streamID: $streamID, period: $period, from: $from, until: $until) {
      entityID
      period
      time
      readOps
      readBytes
      readRecords
      writeOps
      writeBytes
      writeRecords
    }
  }
`;

export const GET_USER_USAGE = gql`
  query GetUserUsage($userID: UUID!, $period: String!, $from: Time!, $until: Time) {
    getUserUsage(userID: $userID, period: $period, from: $from, until: $until) {
      entityID
      period
      time
      readOps
      readBytes
      readRecords
      writeOps
      writeBytes
      writeRecords
      scanOps
      scanBytes
    }
  }
`;
