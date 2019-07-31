import gql from "graphql-tag";

export const QUERY_RECORDS = gql`
  query Records($projectName: String!, $streamName: String!, $keyFields: [String!]!) {
    records(projectName: $projectName, streamName: $streamName, keyFields: $keyFields) @client {
      recordID
      data
      sequenceNumber
    }
  }
`;
