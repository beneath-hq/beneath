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
    ) @client {
      recordID
      data
      sequenceNumber
    }
  }
`;
