const IS_PRODUCTION = process.env.NODE_ENV === "production";

const HTTP_PROTOCOL = IS_PRODUCTION ? "https" : "http";
const WEBSOCKET_PROTOCOL = IS_PRODUCTION ? "wss" : "ws";

const CLIENT_HOST = IS_PRODUCTION ? "beneath.network" : "localhost:3000";
const API_HOST = IS_PRODUCTION ? "api.beneath.network" : "localhost:4000";

const CLIENT_URL = `${HTTP_PROTOCOL}://${CLIENT_HOST}`;
const API_URL = `${HTTP_PROTOCOL}://${API_HOST}`;

module.exports = {
  IS_PRODUCTION,
  HTTP_PROTOCOL,
  WEBSOCKET_PROTOCOL,
  CLIENT_HOST,
  API_HOST,
  CLIENT_URL,
  API_URL,
};
