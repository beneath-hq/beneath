import gql from "graphql-tag";

export const GET_METRICS = gql`
  query GetMetrics($entityKind: EntityKind!, $entityID: UUID!, $period: String!, $from: Time!, $until: Time) {
    getMetrics(entityKind: $entityKind, entityID: $entityID, period: $period, from: $from, until: $until) {
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

export const GET_ORGANIZATION_METRICS = gql`
  query GetOrganizationMetrics($organizationID: UUID!, $period: String!, $from: Time!, $until: Time) {
    getOrganizationMetrics(organizationID: $organizationID, period: $period, from: $from, until: $until) {
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

export const GET_SERVICE_METRICS = gql`
  query GetServiceMetrics($serviceID: UUID!, $period: String!, $from: Time!, $until: Time) {
    getServiceMetrics(serviceID: $serviceID, period: $period, from: $from, until: $until) {
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

export const GET_STREAM_METRICS = gql`
  query GetStreamMetrics($streamID: UUID!, $period: String!, $from: Time!, $until: Time) {
    getStreamMetrics(streamID: $streamID, period: $period, from: $from, until: $until) {
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

export const GET_USER_METRICS = gql`
  query GetUserMetrics($userID: UUID!, $period: String!, $from: Time!, $until: Time) {
    getUserMetrics(userID: $userID, period: $period, from: $from, until: $until) {
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
