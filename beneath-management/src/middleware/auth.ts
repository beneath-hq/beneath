import express from "express";
import _ from "lodash";
import passport from "passport";
import { Strategy as AnonymousStrategy } from "passport-anonymous";
import { Strategy as GitHubStrategy } from "passport-github2";
import { OAuth2Strategy as GoogleStrategy } from "passport-google-oauth";
import { Strategy as BearerStrategy } from "passport-http-bearer";

import { Key, KeyRole } from "../entities/Key";
import { User } from "../entities/User";
import logger from "../lib/logger";
import { IAuthenticatedRequest } from "../types";

const successRedirect = `${process.env.CLIENT_HOST}/auth/callback/login`;
const failureRedirect = `${process.env.CLIENT_HOST}/auth`;

export const apply = (app: express.Express) => {
  // config
  app.use(passport.initialize());

  // apply strategies
  applyGithub(app);
  applyGoogle(app);
  applyBearer(app);
  passport.use(new AnonymousStrategy());
  app.use(passport.authenticate(["bearer", "anonymous"], { session: false }));

  // Middleware to mark unauthenticated users anonymous
  app.use((req: IAuthenticatedRequest, res, next) => {
    if (!req.user) {
      req.user = {
        anonymous: true,
        key: null,
      };
    }
    next();
  });

  // logout endpoint
  app.get("/auth/logout", (req: IAuthenticatedRequest, res) => {
    if (!req.user.anonymous && req.user.key.role === KeyRole.Manage) {
      logger.info(`Logout user ${JSON.stringify(req.user.key)}`);
      req.user.key.remove();
    } else {
      logger.error(`Can't logout user: ${JSON.stringify(req.user)} (Authorization: ${req.header("Authorization")})`);
    }
    res.status(200);
    res.end();
  });
};

export default { apply };

const applyBearer = (app: express.Express) => {
  passport.use(new BearerStrategy(async (token: string, done: any) => {
    try {
      const key = await Key.authenticateKey(token);
      if (key) {
        done(null, { key, anonymous: false });
      } else {
        done(null, false);
      }
    } catch (err) {
      done(err, null);
    }
  }));
};

const applyGithub = (app: express.Express) => {
  const options = {
    callbackURL: process.env.GITHUB_CALLBACK_URL,
    clientID: process.env.GITHUB_CLIENT_ID,
    clientSecret: process.env.GITHUB_CLIENT_SECRET,
    scope: "user:email",
  };

  passport.use(new GitHubStrategy(options, (accessToken: any, refreshToken: any, profile: any, done: any) => {
    handleProfile("github", profile, done);
    logger.info(`Login with Github ID <${profile.id}>`);
  }));

  app.get("/auth/github", passport.authenticate("github"));
  app.get(
    "/auth/github/callback",
    passport.authenticate("github", { session: false, failureRedirect }),
    handleSuccessRedirect
  );
};

const applyGoogle = (app: express.Express) => {
  const options = {
    callbackURL: process.env.GOOGLE_CALLBACK_URL,
    clientID: process.env.GOOGLE_CLIENT_ID,
    clientSecret: process.env.GOOGLE_CLIENT_SECRET,
  };

  passport.use(new GoogleStrategy(options, (accessToken: any, refreshToken: any, profile: any, done: any) => {
    handleProfile("google", profile, done);
    logger.info(`Login with Google ID <${profile.id}>`);
  }));

  const scope = ["https://www.googleapis.com/auth/plus.login", "email"];
  app.get("/auth/google", passport.authenticate("google", { scope }));
  app.get(
    "/auth/google/callback",
    passport.authenticate("google", { session: false, failureRedirect }),
    handleSuccessRedirect
  );
};

const handleSuccessRedirect = (req: IAuthenticatedRequest, res) => {
  const token = req.user.key.keyString;
  res.redirect(`${successRedirect}/?token=${token}`);
};

const handleProfile = async (serviceName: "github"|"google", profile: any, done: any) => {
  try {
    // read profile data
    const id = profile.id;
    const name = profile.displayName;
    const photoUrl = _.get(profile, "photos[0].value");
    let primaryEmail = _.get(profile, "emails[0].value");
    for (const email of profile.emails) { // the first email isn't necessarily the primary
      if (email.primary) {
        primaryEmail = email.value;
        break;
      }
    }

    // create/update/get user
    const user = await User.createOrUpdate({
      email: primaryEmail,
      githubId: serviceName === "github" ? id : undefined,
      googleId: serviceName === "google" ? id : undefined,
      name,
      photoUrl,
    });

    // done
    const key = await Key.issueUserKey(user.userId, KeyRole.Manage, `Browser session`);
    done(undefined, {
      anonymous: false,
      key,
    });
  } catch (err) {
    done(err, undefined);
  }
};
