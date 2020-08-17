import { ApolloClient, ApolloLink, HttpLink, InMemoryCache, defaultDataIdFromObject } from "@apollo/client";
import { onError } from "@apollo/client/link/error";

import fetch from "isomorphic-unfetch";

import { API_URL, IS_PRODUCTION } from "../lib/connection";
import possibleTypes from "./possibleTypes.json";
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
  const typePolicies = {
    // Data normalization config. See https://www.apollographql.com/docs/react/caching/cache-configuration/#data-normalization
    // NOTE: setting keyFields to false causes objects to be embedded in the entry of their parent object, see: https://www.apollographql.com/docs/react/caching/cache-configuration/#disabling-normalization
    BillingInfo: { keyFields: ["organizationID"] },
    BillingMethod: { keyFields: ["billingMethodID"] },
    BillingPlan: { keyFields: ["billingPlanID"] },
    Metrics: { keyFields: ["entityID", "period", "time"] },
    NewUserSecret: { keyFields: ["secretString"] },
    Organization: { keyFields: ["organizationID"] },
    OrganizationMember: { keyFields: ["organizationID", "userID"] },
    PermissionsServicesStreams: { keyFields: false },
    PermissionsUsersOrganizations: { keyFields: false },
    PermissionsUsersProjects: { keyFields: false },
    PrivateOrganization: { keyFields: ["organizationID"] },
    PrivateUser: { keyFields: ["userID"] },
    Project: { keyFields: ["projectID"] },
    ProjectMember: { keyFields: ["projectID", "userID"] },
    PublicOrganization: { keyFields: ["organizationID"] },
    Service: { keyFields: ["serviceID"] },
    Stream: { keyFields: ["streamID"] },
    StreamIndex: { keyFields: ["indexID"] },
    StreamInstance: { keyFields: ["streamInstanceID"] },
    UserSecret: { keyFields: ["userSecretID"] },
  };

  const cache = new InMemoryCache({ possibleTypes, typePolicies }).restore(initialState || {});

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
    link: ApolloLink.from([onError(makeErrorHook({ token, res })), new HttpLink(linkOptions)]),
    cache,
    typeDefs,
    resolvers,
    defaultOptions: {
      mutate: {
        errorPolicy: "all",
      },
      query: {
        errorPolicy: "all",
      },
    },
  });
};

const makeErrorHook = ({ token, res }) => {
  return ({ graphQLErrors, networkError }) => {
    // redirect to /auth/logout if error is `unauthenticated` (probably means the user logged out in another window)
    if (networkError?.result?.error?.match(/authentication error.*/)) {
      if (token) {
        if (process.browser) {
          document.location.href = "/-/redirects/auth/logout";
        } else {
          res.redirect("/-/redirects/auth/logout");
        }
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
