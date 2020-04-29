const next = require("next");
const express = require("express");
const connection = require("./lib/connection");
const cookieParser = require("cookie-parser");
const cors = require("cors");
const fetch = require("isomorphic-unfetch");
const helmet = require("helmet");
const compression = require("compression");
const redirectToHTTPS = require("express-http-to-https").redirectToHTTPS;
const uuidv4 = require("uuid/v4");

global.fetch = require("isomorphic-unfetch"); // Polyfill fetch() on the server (used by apollo-client)

const port = parseInt(process.env.PORT, 10) || 3000;
const dev = process.env.NODE_ENV !== "production";
const app = next({ dev });
const handle = app.getRequestHandler();
const cookieAge = 20 * 365 * 24 * 60 * 60 * 1000;

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

  // Parse cookies for logout route
  server.use(cookieParser());

  // Set anonymous id
  server.use((req, res, next) => {
    const aid = req.cookies.aid;
    if (!aid) {
      res.cookie("aid", uuidv4(), { maxAge: cookieAge, secure: !dev });
    }
    next();
  });

  // Redirect "/" based on whether user logged in
  server.get("/", (req, res, next) => {
    let loggedIn = !!req.cookies.token;
    if (!loggedIn) {
      let noredirect = req.query.noredirect;
      if (!noredirect) {
        res.redirect("https://about.beneath.dev/");
        return;
      }
    }
    next();
  });

  // Redirect "/-/" to "/"
  server.get("/-/", (req, res) => {
    res.redirect("/");
  });

  // Redirected to by backend after login
  server.get("/-/redirects/auth/login/callback", (req, res) => {
    let token = req.query.token;
    if (token) {
      res.cookie("token", token, { secure: !dev });
    } else {
      res.clearCookie("token");
    }
    res.redirect("/");
  });

  // Logout
  server.get("/-/redirects/auth/logout", (req, res) => {
    // If the user is logged in, we'll let the backend logout the token (i.e. delete the token from it's registries)
    // Can't do it with a direct <a> on the client because we have to set the token in the header
    let token = req.cookies.token;
    if (token) {
      let headers = { authorization: `Bearer ${token}` };
      fetch(`${connection.API_URL}/auth/logout`, { headers }).then((r) => {
        console.log(`Successfully logged out ${token}`);
      }).catch((e) => {
        console.error("Error occurred calling backend /auth/logout: ", e);
      });
    }
    res.clearCookie("token");
    res.redirect("/");
  });

  // Redirect to user's secrets
  server.get("/-/redirects/secrets", (req, res) => {
    let loggedIn = !!req.cookies.token;
    if (loggedIn) {
      // "me" gets replaced with billing org name in the organization page
      res.redirect("/me/-/secrets");
    } else {
      res.redirect("/-/auth");
    }
  });

  // Redirect to billing
  server.get("/-/redirects/upgrade-pro", (req, res) => {
    let loggedIn = !!req.cookies.token;
    if (loggedIn) {
      // "me" gets replaced with billing org name in the organization page
      res.redirect("/me/-/billing");
    } else {
      res.redirect("/-/auth");
    }
  });

  // Use to config routes
  const addStaticRoute = (route) => {
    server.get(route, (req, res) => {
      return handle(req, res);
    });
  };

  const addDynamicRoute = (route, page) => {
    server.get(route, (req, res) => {
      app.render(req, res, page, req.params);
    });
  };

  // Add routes in order of precedence
  addStaticRoute("/-/*");
  addStaticRoute("/assets/*");
  addStaticRoute("/robots.txt");
  addStaticRoute("/favicon.ico");
  addDynamicRoute("/:organization_name", "/organization");
  addDynamicRoute("/:organization_name/-/:tab", "/organization");
  addDynamicRoute("/:organization_name/:project_name", "/project");
  addDynamicRoute("/:organization_name/:project_name/-/:tab", "/project");
  addDynamicRoute("/:organization_name/:project_name/streams/:stream_name", "/stream");
  addDynamicRoute("/:organization_name/:project_name/streams/:stream_name/-/:tab", "/stream");
  addStaticRoute("*"); // catchall

  // Run server
  server.listen(port, err => {
    if (err) {
      throw err;
    }
    console.log(`Ready on ${port}`);
  });
});
