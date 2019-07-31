import gql from "graphql-tag";

export const QUERY_RECORDS = gql`
  query Records($instanceID: UUID!) {
    records(instanceID: $instanceID) @client {
      recordID
      data
      sequenceNumber
    }
  }
`;
