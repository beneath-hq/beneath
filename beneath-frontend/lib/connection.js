export const IS_PRODUCTION = process.env.NODE_ENV === "production";

export const HTTP_PROTOCOL = IS_PRODUCTION ? "https" : "http";
export const WEBSOCKET_PROTOCOL = IS_PRODUCTION ? "wss" : "ws";

export const CLIENT_HOST = IS_PRODUCTION ? "beneath.network" : "localhost:3000";
export const API_HOST = IS_PRODUCTION ? "api.beneath.network" : "localhost:4000";

export const CLIENT_URL = `${HTTP_PROTOCOL}://${CLIENT_HOST}`;
export const API_URL = `${HTTP_PROTOCOL}://${API_HOST}`;

export default {
  IS_PRODUCTION,
  HTTP_PROTOCOL,
  WEBSOCKET_PROTOCOL,
  CLIENT_HOST,
  API_HOST,
  CLIENT_URL,
  API_URL,
};
