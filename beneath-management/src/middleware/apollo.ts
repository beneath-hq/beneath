import { ApolloServer } from "apollo-server-express";
import express from "express";
import { GraphQLError } from "graphql";

import logger from "../lib/logger";
import { resolvers, typeDefs } from "../schema";
import { IAuthenticatedRequest } from "../types";

export const apply = (app: express.Express) => {
  const path = "/graphql";
  const server = new ApolloServer({
    typeDefs,
    resolvers,
    formatError: (error: GraphQLError) => {
      logger.error(error);
      return error;
    },
    context: ({ req }: { req: IAuthenticatedRequest }) => {
      return {
        user: req.user
      };
    },
    introspection: true,
    playground: {
      settings: {
        "request.credentials": "include",
      },
    },
  });
  server.applyMiddleware({ app, path });
};

export default { apply };
