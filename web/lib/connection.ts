const IS_PRODUCTION = process.env.NODE_ENV === "production";

const HTTP_PROTOCOL = IS_PRODUCTION ? "https" : "http";
const WEBSOCKET_PROTOCOL = IS_PRODUCTION ? "wss" : "ws";

const CLIENT_HOST = IS_PRODUCTION ? "beneath.dev" : "localhost:3000";
const CLIENT_URL = `${HTTP_PROTOCOL}://${CLIENT_HOST}`;

const API_HOST = IS_PRODUCTION ? "control.beneath.dev" : "localhost:4000";
const API_URL = `${HTTP_PROTOCOL}://${API_HOST}`;

const GATEWAY_HOST = IS_PRODUCTION ? "data.beneath.dev" : "localhost:5000";
const GATEWAY_URL = `${HTTP_PROTOCOL}://${GATEWAY_HOST}`;
const GATEWAY_URL_WS = `${WEBSOCKET_PROTOCOL}://${GATEWAY_HOST}`;

export default {
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
};
