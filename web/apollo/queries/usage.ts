import gql from "graphql-tag";

export const GET_USAGE = gql`
  query GetUsage($input: GetUsageInput!) {
    getUsage(input: $input) {
      entityID
      label
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
  query GetOrganizationUsage($input: GetEntityUsageInput!) {
    getOrganizationUsage(input: $input) {
      entityID
      label
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
  query GetServiceUsage($input: GetEntityUsageInput!) {
    getServiceUsage(input: $input) {
      entityID
      label
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
  query GetStreamUsage($input: GetEntityUsageInput!) {
    getStreamUsage(input: $input) {
      entityID
      label
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
  query GetUserUsage($input: GetEntityUsageInput!) {
    getUserUsage(input: $input) {
      entityID
      label
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
