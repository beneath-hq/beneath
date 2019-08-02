import gql from "graphql-tag";

export const QUERY_RECORDS = gql`
  query Records(
    $projectName: String!,
    $streamName: String!,
    $keyFields: [String!]!,
    $limit: Int!,
    $where: JSON)
  {
    records(
      projectName: $projectName,
      streamName: $streamName,
      keyFields: $keyFields,
      limit: $limit,
      where: $where
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

export const CREATE_RECORDS = gql`
  mutation CreateRecords($instanceID: UUID!, $json: JSON!) {
    createRecords(instanceID: $instanceID, json: $json) @client {
      error
    }
  }
`;
