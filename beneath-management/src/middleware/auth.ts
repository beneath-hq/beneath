import express from "express";
import _ from "lodash";
import passport from "passport";
import { Strategy as GitHubStrategy } from "passport-github2";
import { OAuth2Strategy as GoogleStrategy } from "passport-google-oauth";

import { User } from "../entities/User";
import logger from "../lib/logger";
import { IAuthenticatedRequest } from "../types";

const successRedirect = `${process.env.CLIENT_HOST}/`;
const failureRedirect = `${process.env.CLIENT_HOST}/auth`;
const logoutRedirect = `${process.env.CLIENT_HOST}/`;

export const apply = (app: express.Express) => {
  passport.serializeUser(serializeUser);
  passport.deserializeUser(deserializeUser);
  app.use(passport.initialize());
  applyGithub(app);
  applyGoogle(app);
  app.use(passport.session());
  app.get("/auth/logout", (req: IAuthenticatedRequest, res) => {
    logger.info(`Logout user ${JSON.stringify(req.user)}`);
    req.logout();
    res.redirect(logoutRedirect);
  });
};

export default { apply };

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
  app.get("/auth/github/callback", passport.authenticate("github", { successRedirect, failureRedirect }));
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
  app.get("/auth/google/callback", passport.authenticate("google", { successRedirect, failureRedirect }));
};

const serializeUser = (user, done) => {
  done(undefined, user.userId);
};

const deserializeUser = (userId, done) => {
  done(undefined, { userId, kind: "session" });
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
    done(undefined, {
      userId: user.userId,
      kind: "session",
    });
  } catch (err) {
    done(err, undefined);
  }
};
