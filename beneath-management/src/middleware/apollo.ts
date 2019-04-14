import { ApolloServer } from "apollo-server-express";
import express from "express";
import { GraphQLError } from "graphql";
import graphqlDepthLimit from "graphql-depth-limit";

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
    tracing: process.env.NODE_ENV !== "production",
    playground: {
      settings: {
        "request.credentials": "include",
      },
    },
    validationRules: [
      graphqlDepthLimit(3)
    ],
  });
  server.applyMiddleware({ app, path });
};

export default { apply };
