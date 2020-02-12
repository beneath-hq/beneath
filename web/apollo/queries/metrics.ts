import gql from "graphql-tag";

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
    }
  }
`;
