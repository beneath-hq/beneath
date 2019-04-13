import express from "express";
import helmet from "helmet";
import { redirectToHTTPS } from "express-http-to-https";

export const apply = (app: express.Express) => {
  // force https
  app.enable("trust proxy");
  app.use(redirectToHTTPS([/localhost:(\d+)/], [], 301));

  // headers security
  app.use(helmet());
};

export default { apply };
