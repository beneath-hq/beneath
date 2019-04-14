import { ApolloServer } from "apollo-server-express";
import express from "express";
import { GraphQLError } from "graphql";

import logger from "../lib/logger";
import { resolvers, typeDefs } from "../schema";

export const apply = (app: express.Express) => {
  const path = "/graphql";
  const server = new ApolloServer({
    typeDefs,
    resolvers,
    formatError: (error: GraphQLError) => {
      logger.error(error);
      return error;
    },
  });
  server.applyMiddleware({ app, path });
};

export default { apply };
