import gql from "graphql-tag";

export const GET_AID = gql`
  query AID {
    aid @client
  }
`;

export const GET_TOKEN = gql`
  query Token {
    token @client
  }
`;
