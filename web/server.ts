import next from "next";
import express from "express";
import { API_URL } from "./lib/connection";
import cookieParser from "cookie-parser";
import cors from "cors";
import fetch from "isomorphic-unfetch";
import helmet from "helmet";
import compression from "compression";
import { redirectToHTTPS } from "express-http-to-https";
import { v4 as uuidv4 } from "uuid";

import "isomorphic-unfetch"; // Polyfill fetch() on the server (used by apollo-client)

const port = process.env.PORT ? parseInt(process.env.PORT, 10) : 3000;
const dev = process.env.NODE_ENV !== "production";
const app = next({ dev });
const handle = app.getRequestHandler();
const cookieAge = 20 * 365 * 24 * 60 * 60 * 1000;

app.prepare().then(() => {
  const server = express();

  // Health check
  server.get("/healthz", (_, res) => {
    res.sendStatus(200);
  });

  // Redirect https
  server.enable("trust proxy");
  server.use(redirectToHTTPS([/localhost:(\d+)/], [], 301));

  // Enable compression
  server.use(compression());

  // Headers security
  // server.use(helmet.contentSecurityPolicy());
  server.use(helmet.dnsPrefetchControl());
  server.use(helmet.expectCt());
  server.use(helmet.frameguard());
  server.use(helmet.hidePoweredBy());
  server.use(helmet.hsts());
  server.use(helmet.ieNoOpen());
  server.use(helmet.noSniff());
  server.use(helmet.permittedCrossDomainPolicies());
  server.use(helmet.referrerPolicy());
  server.use(helmet.xssFilter());

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

  // Redirect "/" based on whether user logged in (only in production)
  if (!dev) {
    server.get("/", (req, res, next) => {
      const loggedIn = !!req.cookies.token;
      if (!loggedIn) {
        const noredirect = req.query.noredirect;
        if (!noredirect) {
          res.redirect("https://about.beneath.dev/");
          return;
        }
      }
      next();
    });
  }

  // Redirect "/-/" to "/"
  server.get("/-/", (_, res) => {
    res.redirect("/");
  });

  // Redirected to by backend after login
  server.get("/-/redirects/auth/login/callback", (req, res) => {
    const token = req.query.token;
    if (token) {
      res.cookie("token", token, { maxAge: cookieAge, secure: !dev });
    } else {
      res.clearCookie("token");
    }
    res.redirect("/");
  });

  // Logout
  server.get("/-/redirects/auth/logout", (req, res) => {
    // If the user is logged in, we'll let the backend logout the token (i.e. delete the token from it's registries)
    // Can't do it with a direct <a> on the client because we have to set the token in the header
    const token = req.cookies.token;
    if (token) {
      const headers = { authorization: `Bearer ${token}` };
      fetch(`${API_URL}/auth/logout`, { headers })
        .then((_) => {
          console.log(`Successfully logged out ${token}`);
        })
        .catch((e) => {
          console.error("Error occurred calling backend /auth/logout: ", e);
        });
    }
    res.clearCookie("token");
    res.redirect("/");
  });

  // Redirect to user's secrets
  server.get("/-/redirects/secrets", (req, res) => {
    const loggedIn = !!req.cookies.token;
    if (loggedIn) {
      // "me" gets replaced with billing org name in the organization page
      res.redirect("/me/-/secrets");
    } else {
      res.redirect("/-/auth");
    }
  });

  // Redirect to billing
  server.get("/-/redirects/upgrade-pro", (req, res) => {
    const loggedIn = !!req.cookies.token;
    if (loggedIn) {
      // "me" gets replaced with billing org name in the organization page
      res.redirect("/me/-/billing");
    } else {
      res.redirect("/-/auth");
    }
  });

  // Use to config routes
  const addStaticRoute = (route: string) => {
    server.get(route, (req, res) => {
      return handle(req, res);
    });
  };

  const addDynamicRoute = (route: string, page: string) => {
    server.get(route, (req, res) => {
      const params = { ...req.params };

      // add query args to route params
      for (const key in req.query) {
        const value = req.query[key]
        if (typeof value === "string") {
          params[key] = value;
        }
      }

      app.render(req, res, page, params);
    });
  };

  // Add routes in order of precedence
  addStaticRoute("/-/*");
  addStaticRoute("/assets/*");
  addStaticRoute("/robots.txt");
  addStaticRoute("/favicon.ico");
  addDynamicRoute("/:organization_name", "/organization");
  addDynamicRoute("/:organization_name/-/:tab", "/organization");
  addDynamicRoute("/:organization_name/-/billing/checkout", "/-/billing/checkout");
  addDynamicRoute("/:organization_name/:project_name", "/project");
  addDynamicRoute("/:organization_name/:project_name/-/:tab", "/project");
  addDynamicRoute("/:organization_name/:project_name/service::service_name", "/service");
  addDynamicRoute("/:organization_name/:project_name/service::service_name/-/:tab", "/service");
  addDynamicRoute("/:organization_name/:project_name/stream::stream_name", "/stream");
  addDynamicRoute("/:organization_name/:project_name/stream::stream_name/-/:tab", "/stream");
  addDynamicRoute("/:organization_name/:project_name/stream::stream_name/:version", "/stream");
  addDynamicRoute("/:organization_name/:project_name/stream::stream_name/:version/-/:tab", "/stream");
  addStaticRoute("*"); // catchall

  // Run server
  server.listen(port, () => {
    console.log(`Ready on ${port}`);
  });
});
