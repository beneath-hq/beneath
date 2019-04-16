import redis from "redis";
import logger from "./logger";

const url = process.env.REDIS_URL;
let client = null;
if (url) {
  client = redis.createClient(url);
  client.on("ready", () => {
    logger.info("Redis is running");
  });
  client.on("error", (err: any) => {
    logger.error("Redis error", err);
  });
} else {
  logger.info(`Running without Redis`);
}

export default client;
