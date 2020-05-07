import gql from "graphql-tag";

export const CREATE_RECORDS = gql`
  mutation CreateRecords($instanceID: UUID!, $json: JSON!) {
    createRecords(instanceID: $instanceID, json: $json) @client {
      error
    }
  }
`;
