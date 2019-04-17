import cors from "cors";
import express from "express";
import { redirectToHTTPS } from "express-http-to-https";
import helmet from "helmet";

export const apply = (app: express.Express) => {
  // force https
  app.enable("trust proxy");
  app.use(redirectToHTTPS([/localhost:(\d+)/], [], 301));

  // headers security
  app.use(cors());
  app.use(helmet());
};

export default { apply };
