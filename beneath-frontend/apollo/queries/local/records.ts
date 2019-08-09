import gql from "graphql-tag";

export const QUERY_RECORDS = gql`
  query Records(
    $projectName: String!,
    $streamName: String!,
    $limit: Int!,
    $where: JSON,
    $after: JSON)
  {
    records(
      projectName: $projectName,
      streamName: $streamName,
      limit: $limit,
      where: $where,
      after: $after
    ) @client @connection(key: "records", filter: ["projectName", "streamName"]) {
      data {
        recordID
        data
        sequenceNumber
      }
      error
    }
  }
`;

export const QUERY_LATEST_RECORDS = gql`
  query LatestRecords(
    $projectName: String!,
    $streamName: String!,
    $limit: Int!,
  ) {
    latestRecords(
      projectName: $projectName,
      streamName: $streamName,
      limit: $limit,
    ) @client @connection(key: "latestRecords", filter: ["projectName", "streamName"]) {
      recordID
      data
      sequenceNumber
    }
  }
`;

export const CREATE_RECORDS = gql`
  mutation CreateRecords($instanceID: UUID!, $json: JSON!) {
    createRecords(instanceID: $instanceID, json: $json) @client {
      error
    }
  }
`;
