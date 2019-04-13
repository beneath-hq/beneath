import redis from "redis";
import logger from "./logger";

const url = process.env.REDIS_URL;
const client = redis.createClient(url);

client.on("error", (err: any) => {
  logger.error("Redis error", err);
});

export default client;
