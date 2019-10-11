import gql from "graphql-tag";

import { GET_AID, GET_TOKEN } from "../queries/local/token";

export const typeDefs = gql`
  extend type Query {
    aid: String
    token: String
  }
`;

export const resolvers = {
  Query: {
    aid: async (_: any, args: any, { cache }: any) => {
      const { aid } = cache.readQuery({ query: GET_AID });
      return aid;
    },
    token: async (_: any, args: any, { cache }: any) => {
      const { token } = cache.readQuery({ query: GET_TOKEN });
      return token;
    },
  },
};

export default {
  typeDefs,
  resolvers,
};
