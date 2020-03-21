export const ENV = (typeof process !== "undefined") && process.env.BENEATH_ENV || "production";

export const DEV = ["dev", "development"].includes(ENV);

export const JS_CLIENT_ID = "beneath-js";

export const BENEATH_FRONTEND_HOST = DEV ? "http://localhost:3000" : "https://beneath.dev";

export const BENEATH_CONTROL_HOST = DEV ? "http://localhost:4000" : "https://control.beneath.network";

export const BENEATH_GATEWAY_HOST = DEV ? "http://localhost:5000" : "https://data.beneath.network";
