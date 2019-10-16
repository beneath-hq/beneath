import { ApolloClient, ApolloLink, HttpLink, InMemoryCache, defaultDataIdFromObject } from "apollo-boost";
import { ErrorLink } from "apollo-link-error";

import { API_URL, IS_PRODUCTION } from "../lib/connection";
import { resolvers, typeDefs } from "./schema";
import { GET_AID, GET_TOKEN } from "./queries/local/token";

let apolloClient = null;

export const getApolloClient = (options) => {
  // Creates new client for every server-side request (to avoid bad data sharing)
  if (!process.browser) {
    return createApolloClient(options);
  }

  // Reuse client on the client-side
  if (!apolloClient) {
    apolloClient = createApolloClient(options);
  }

  return apolloClient;
};

const createApolloClient = ({ initialState, anonymousID, token, res }) => {
  const linkOptions = {
    uri: `${API_URL}/graphql`,
    credentials: "include",
  };

  linkOptions.headers = {};
  if (token) {
    linkOptions.headers.Authorization = `Bearer ${token}`;
  }
  if (anonymousID) {
    linkOptions.headers["X-Beneath-Aid"] = anonymousID;
  }

  const cache = new InMemoryCache({ dataIdFromObject }).restore(initialState || {});
  cache.writeQuery({ query: GET_AID, data: { aid: anonymousID } });
  cache.writeQuery({ query: GET_TOKEN, data: { token } });

  const apolloOptions = {
    connectToDevTools: process.browser && !IS_PRODUCTION,
    ssrMode: !process.browser,
    link: ApolloLink.from([new ErrorLink(makeErrorHook({ token, res })), new HttpLink(linkOptions)]),
    cache,
    typeDefs,
    resolvers,
  };

  return new ApolloClient(apolloOptions);
};

// returns ID for objects for caching/automatic state update
const dataIdFromObject = (object) => {
  switch (object.__typename) {
    case "User":
      return object.userID;
    case "Me":
      return `me:${object.userID}`;
    case "Secret":
      return `${object.secretID}`;
    case "NewSecret":
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
    if (networkError && networkError.result && networkError.result.error === "unauthenticated") {
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
