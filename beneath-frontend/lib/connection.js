export const IS_PRODUCTION = process.env.NODE_ENV === "production";

export const BASE_URL = IS_PRODUCTION
  ? `https://beneath.network`
  : `http://localhost:3000`;

export const API_HOST = IS_PRODUCTION
  ? "api.beneath.network"
  : "localhost:4000";

export const WEBSOCKET_PROTOCOL = IS_PRODUCTION ? "wss" : "ws";

export const HTTP_PROTOCOL = IS_PRODUCTION ? "https" : "http";

export default {
  IS_PRODUCTION,
  BASE_URL,
  API_HOST,
  WEBSOCKET_PROTOCOL,
  HTTP_PROTOCOL,
};
