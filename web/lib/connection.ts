export const IS_PRODUCTION = process.env.NODE_ENV === "production";

export const HTTP_PROTOCOL = IS_PRODUCTION ? "https" : "http";
export const WEBSOCKET_PROTOCOL = IS_PRODUCTION ? "wss" : "ws";

export const CLIENT_HOST = IS_PRODUCTION ? "beneath.dev" : "localhost:3000";
export const CLIENT_URL = `${HTTP_PROTOCOL}://${CLIENT_HOST}`;

export const API_HOST = IS_PRODUCTION ? "control.beneath.dev" : "localhost:4000";
export const API_URL = `${HTTP_PROTOCOL}://${API_HOST}`;

export const GATEWAY_HOST = IS_PRODUCTION ? "data.beneath.dev" : "localhost:5000";
export const GATEWAY_URL = `${HTTP_PROTOCOL}://${GATEWAY_HOST}`;
export const GATEWAY_URL_WS = `${WEBSOCKET_PROTOCOL}://${GATEWAY_HOST}`;
