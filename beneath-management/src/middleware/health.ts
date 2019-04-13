import express from "express";
import { getConnection } from "typeorm";

import logger from "../lib/logger";

const healthCheck = (req: express.Request, res: express.Response) => {
  // postgres.data.proc("version").then(...).catch(...)
  if (getConnection().isConnected) {
    res.sendStatus(200);
  } else {
    logger.error(`Health check failed`);
    res.sendStatus(500);
  }
}

export const apply = (app: express.Express) => {
  app.get("/", healthCheck);
  app.get("/healthz", healthCheck);  
};

export default { apply };
