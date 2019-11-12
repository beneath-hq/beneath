import { ApolloClient, ApolloLink, HttpLink, InMemoryCache, defaultDataIdFromObject } from "apollo-boost";
import { ErrorLink } from "apollo-link-error";
import fetch from "isomorphic-unfetch";

import { API_URL, IS_PRODUCTION } from "../lib/connection";
import { resolvers, typeDefs } from "./schema";
import { GET_AID, GET_TOKEN } from "./queries/local/token";

let apolloClient = null;

export const getApolloClient = ({ req, res, initialState }) => {
  // Creates new client for every server-side request (to avoid bad data sharing)
  if (typeof window === "undefined") {
    return createApolloClient({ req, res, initialState });
  }

  // Reuse client on the client-side
  if (!apolloClient) {
    apolloClient = createApolloClient({ req, res, initialState });
  }

  return apolloClient;
};

const createApolloClient = ({ req, res, initialState }) => {
  const cache = new InMemoryCache({ dataIdFromObject }).restore(initialState || {});

  const linkOptions = {
    credentials: "include",
    fetch,
    headers: {},
    uri: `${API_URL}/graphql`,
  };

  // read token from cookie
  const token = readTokenFromCookie(req);
  if (token) {
    linkOptions.headers.Authorization = `Bearer ${token}`;
    cache.writeQuery({ query: GET_TOKEN, data: { token } });
  }

  // read anonymous id from cookie
  const anonymousID = readAnonymousIDFromCookie(req);
  if (anonymousID) {
    linkOptions.headers["X-Beneath-Aid"] = anonymousID;
    cache.writeQuery({ query: GET_AID, data: { aid: anonymousID } });
  }

  return new ApolloClient({
    connectToDevTools: typeof window === "undefined" && !IS_PRODUCTION,
    ssrMode: typeof window === "undefined",
    link: ApolloLink.from([new ErrorLink(makeErrorHook({ token, res })), new HttpLink(linkOptions)]),
    cache,
    typeDefs,
    resolvers,
  });
};

// returns ID for objects for caching/automatic state update
const dataIdFromObject = (object) => {
  switch (object.__typename) {
    case "User":
      return object.userID;
    case "Me":
      return `me:${object.userID}`;
    case "UserSecret":
      return `${object.userSecretID}`;
    case "NewUserSecret":
      return `${object.secretString}`;
    case "Project":
      return `${object.projectID}`;
    case "Stream":
      return `${object.streamID}`;
    case "Record":
      return `${object.recordID}`;
    case "RecordsResponse":
      return defaultDataIdFromObject(object);
    case "CreateRecordsResponse":
      return defaultDataIdFromObject(object);
    case "Metrics":
      return `metrics:${object.entityID}:${object.period}:${object.time}`;
    default: {
      console.warn(`Unknown typename in dataIdFromObject: ${object.__typename}`);
      return defaultDataIdFromObject(object);
    }
  }
};

const makeErrorHook = ({ token, res }) => {
  return ({ graphQLErrors, networkError }) => {
    // redirect to /auth/logout if error is `unauthenticated` (probably means the user logged out in another window)
    if (
      networkError &&
      networkError.result &&
      networkError.result.error.match(/authentication error.*/)
    ) {
      if (token) {
        if (process.browser) {
          document.location.href = "/auth/logout";
        } else {
          res.redirect("/auth/logout");
        }
      }
    }

    if (graphQLErrors && graphQLErrors.length > 0) {
      let error = graphQLErrors[0];
      if (error.extensions && error.extensions.code === "VALIDATION_ERROR") {
        console.log("Validation error", error.extensions.exception);
      }
    }
  };
};

const readTokenFromCookie = (maybeReq) => {
  let token = null;
  let cookie = null;
  if (maybeReq) {
    cookie = maybeReq.headers.cookie;
  } else if (typeof document !== "undefined") {
    cookie = document.cookie;
  }
  if (cookie) {
    cookie = cookie.split(";").find((c) => c.trim().startsWith("token="));
    if (cookie) {
      token = cookie.replace(/^\s*token=/, "");
      token = decodeURIComponent(token);
      if (token.length === 0) {
        token = null;
      }
    }
  }
  return token;
};

const readAnonymousIDFromCookie = (maybeReq) => {
  let aid = null;
  let cookie = null;
  if (maybeReq) {
    cookie = maybeReq.headers.cookie;
  } else if (typeof document !== "undefined") {
    cookie = document.cookie;
  }
  if (cookie) {
    cookie = cookie.split(";").find((c) => c.trim().startsWith("aid="));
    if (cookie) {
      aid = cookie.replace(/^\s*aid=/, "");
      aid = decodeURIComponent(aid);
      if (aid.length === 0) {
        aid = null;
      }
    }
  }
  return aid;
};
