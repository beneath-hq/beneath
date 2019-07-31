import gql from "graphql-tag";

import { GET_TOKEN } from "../queries/local/token";

export const typeDefs = gql`
  extend type Query {
    token: String
  }
`;

export const resolvers = {
  Query: {
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
