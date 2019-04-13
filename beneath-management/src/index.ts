import { ApolloServer } from "apollo-server-express";
import express from "express";
import { GraphQLError } from "graphql";
import "reflect-metadata";

import { createConnection } from "typeorm";

import logger from "./lib/logger";
import { resolvers, typeDefs } from "./schema";
import { Project } from "./entities/Project";
import { User } from "./entities/User";

const path = "/graphql";
const port: number = parseInt(process.env.PORT, 10) || 4000;
const app = express();
// app.use(path, jwtCheck);

const server = new ApolloServer({
  typeDefs,
  resolvers,
  formatError: (error: GraphQLError) => {
    logger.error(error);
    return error;
  },
});
server.applyMiddleware({ app, path });

(async () => {
  logger.info(`Connecting to db`);
  const connection = await createConnection();
  
  // const user = new User();
  // user.username = "bem";
  // user.email = "benjamin@beneath.network";
  // user.name = "Benjamin Egelund-Müller";
  // user.bio = "Founder of Beneath";
  // await user.save();

  // const project = new Project();
  // project.name = "makerdao";
  // project.displayName = "MakerDAO";
  // project.description = "Stablecoin and ecentralised credit system";
  // project.site = "https://makerdao.com";
  // project.users = [user];
  // await project.save();

  app.listen({ port }, () => {
    logger.info(`🚀 Server running on port ${4000}`);
  });

})();