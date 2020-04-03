export const ENV = (typeof process !== "undefined") && process.env.BENEATH_ENV || "production";

export const DEV = ["dev", "development"].includes(ENV);

export const JS_CLIENT_ID = "beneath-js";

export const BENEATH_FRONTEND_HOST = DEV ? "http://localhost:3000" : "https://beneath.dev";

export const BENEATH_CONTROL_HOST = DEV ? "http://localhost:4000" : "https://control.beneath.dev";

export const BENEATH_GATEWAY_HOST = DEV ? "http://localhost:5000" : "https://data.beneath.dev";

export const BENEATH_GATEWAY_HOST_GRPC = DEV ? "localhost:50051" : "grpc.data.beneath.dev";

export const BENEATH_GATEWAY_HOST_WS = DEV ? "ws://localhost:5000" : "wss://data.beneath.dev";

export const DEFAULT_READ_BATCH_SIZE = 1000;

export const DEFAULT_SUBSCRIBE_POLL_AT_MOST_EVERY_MS = 250;
