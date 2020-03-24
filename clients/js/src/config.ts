export const ENV = (typeof process !== "undefined") && process.env.BENEATH_ENV || "production";

export const DEV = ["dev", "development"].includes(ENV);

export const JS_CLIENT_ID = "beneath-js";

export const BENEATH_FRONTEND_HOST = DEV ? "http://localhost:3000" : "https://beneath.dev";

export const BENEATH_CONTROL_HOST = DEV ? "http://localhost:4000" : "https://control.beneath.dev";

export const BENEATH_GATEWAY_HOST = DEV ? "http://localhost:5000" : "https://data.beneath.dev";

export const BENEATH_GATEWAY_HOST_GRPC = DEV ? "localhost:50051" : "grpc.data.beneath.dev";

export const DEFAULT_WRITE_BATCH_SIZE = 10000;

export const DEFAULT_WRITE_BATCH_BYTES = 10000000;

export const DEFAULT_WRITE_DELAY_SECONDS = 1.0;

export const DEFAULT_READ_BATCH_SIZE = 1000;

export const DEFAULT_READ_ALL_MAX_BYTES = 10 * 2 ** 20;

export const DEFAULT_SUBSCRIBE_PREFETCHED_RECORDS = 10000;

export const DEFAULT_SUBSCRIBE_CONCURRENT_CALLBACKS = 1000;

export const DEFAULT_SUBSCRIBE_POLL_AT_LEAST_EVERY_SECONDS = 60.0;

export const DEFAULT_SUBSCRIBE_POLL_AT_MOST_EVERY_SECONDS = 1.0;
