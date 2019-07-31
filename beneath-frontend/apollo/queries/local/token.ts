import gql from "graphql-tag";

export const GET_TOKEN = gql`
  query Token {
    token @client
  }
`;
