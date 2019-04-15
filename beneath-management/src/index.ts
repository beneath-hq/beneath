import "reflect-metadata";

import express from "express";
import { createConnection } from "typeorm";

import logger from "./lib/logger";

import apollo from "./middleware/apollo";
import auth from "./middleware/auth";
import health from "./middleware/health";
import security from "./middleware/security";

const port: number = parseInt(process.env.PORT, 10) || 4000;
const app = express();

health.apply(app);
security.apply(app);
auth.apply(app);
apollo.apply(app);

(async () => {
  logger.info(`Connecting to db`);
  await createConnection();
  app.listen({ port }, () => {
    logger.info(`🚀 Server running on port ${4000}`);
  });

})();
