const IS_PRODUCTION = process.env.NODE_ENV === "production";

const HTTP_PROTOCOL = IS_PRODUCTION ? "https" : "http";
const WEBSOCKET_PROTOCOL = IS_PRODUCTION ? "wss" : "ws";

const CLIENT_HOST = IS_PRODUCTION ? "beneath.network" : "localhost:3000";
const CLIENT_URL = `${HTTP_PROTOCOL}://${CLIENT_HOST}`;

const API_HOST = IS_PRODUCTION ? "control.beneath.network" : "localhost:4000";
const API_URL = `${HTTP_PROTOCOL}://${API_HOST}`;

const GATEWAY_HOST = IS_PRODUCTION ? "data.beneath.network" : "localhost:5000";
const GATEWAY_URL = `${HTTP_PROTOCOL}://${GATEWAY_HOST}`;
const GATEWAY_URL_WS = `${WEBSOCKET_PROTOCOL}://${GATEWAY_HOST}`;

const SEGMENT_WRITE_KEY = IS_PRODUCTION ? "ZmE8vve82Ei12garfnkyTLP1PBHgIxBj" : "k7toJbXkajZ8fdJzMQPJwCkqs0rColn9";

module.exports = {
  IS_PRODUCTION,
  HTTP_PROTOCOL,
  WEBSOCKET_PROTOCOL,
  CLIENT_HOST,
  API_HOST,
  CLIENT_URL,
  API_URL,
  GATEWAY_HOST,
  GATEWAY_URL,
  GATEWAY_URL_WS,
  SEGMENT_WRITE_KEY,
};
