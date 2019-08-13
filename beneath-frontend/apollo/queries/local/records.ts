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
        timestamp
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
    $before: Int,
  ) {
    latestRecords(
      projectName: $projectName,
      streamName: $streamName,
      limit: $limit,
      before: $before,
    ) @client @connection(key: "latestRecords", filter: ["projectName", "streamName"]) {
      data {
        recordID
        data
        timestamp
      }
      error
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
