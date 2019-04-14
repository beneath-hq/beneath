import express from "express";
import session from "express-session";

import connectRedis from "connect-redis";
const RedisStore = connectRedis(session);

import redis from "../lib/redis";

export const apply = (app: express.Express) => {
  app.use(session({
    secret: "nrh327fh289fhd238hd0931alo10nvuw",
    cookie: {
      secure: process.env.NODE_ENV === "production",
    },
    resave: false,
    saveUninitialized: false,
    store: new RedisStore({ client: redis, ttl: 604800 }),
  }));
};

export default { apply };
