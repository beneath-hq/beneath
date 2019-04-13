import winston from "winston";

// Get log level
let logLevel = "info";
if ((process.env.NODE_ENV !== "production") || (process.env.DEBUG === "1")) {
  logLevel = "debug";
}

// Setup logger
const logger = winston.createLogger({
  format: winston.format.combine(
    winston.format.timestamp(),
    // winston.format.json()
    winston.format.printf((info: any) => {
      return `${info.timestamp} [${info.level}]: ${info.message}`;
    })
  ),
  level: logLevel,
  exitOnError: false,
  transports: [
    new winston.transports.Console({
      handleExceptions: true,
    }),
  ]
});

// exports
export default logger;
