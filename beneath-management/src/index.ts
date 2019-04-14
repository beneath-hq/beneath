import "reflect-metadata";

import express from "express";
import { createConnection } from "typeorm";

import logger from "./lib/logger";

import apollo from "./middleware/apollo";
import auth from "./middleware/auth";
import health from "./middleware/health";
import security from "./middleware/security";
import session from "./middleware/session";

const port: number = parseInt(process.env.PORT, 10) || 4000;
const app = express();

health.apply(app);
security.apply(app);
session.apply(app);
auth.apply(app);
apollo.apply(app);

(async () => {
  logger.info(`Connecting to db`);
  const connection = await createConnection();

  // import { Project } from "./entities/Project";
  // import { User } from "./entities/User";

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
