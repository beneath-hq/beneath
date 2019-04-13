const next = require("next");
const express = require("express");
const cors = require("cors");
const helmet = require("helmet");
const compression = require("compression");
const redirectToHTTPS = require("express-http-to-https").redirectToHTTPS;

const port = parseInt(process.env.PORT, 10) || 3000;
const dev = process.env.NODE_ENV !== "production";
const app = next({ dev });
const handle = app.getRequestHandler();

app.prepare().then(() => {
  const server = express();

  // Health check
  server.get("/healthz", (req, res) => {
    res.sendStatus(200);
  });

  // Redirect https
  server.enable("trust proxy");
  server.use(redirectToHTTPS([/localhost:(\d+)/], [], 301));

  // Enable compression
  server.use(compression());

  // Headers security
  server.use(helmet());

  // CORS
  server.use(cors());

  // Next.js handlers
  server.get("*", (req, res) => {
    return handle(req, res);
  });

  // Run server
  server.listen(port, err => {
    if (err) throw err;
    console.log(`Ready on ${port}`);
  });
});
