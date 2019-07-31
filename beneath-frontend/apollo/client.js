import { ApolloClient, ApolloLink, HttpLink, InMemoryCache, defaultDataIdFromObject } from "apollo-boost";
import { ErrorLink } from "apollo-link-error";

import { API_URL, IS_PRODUCTION } from "../lib/connection";
import { resolvers, typeDefs } from "./schema";

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

const createApolloClient = ({ initialState, token, res }) => {
  const linkOptions = {
    uri: `${API_URL}/graphql`,
    credentials: "include",
  };

  if (token) {
    linkOptions.headers = {
      Authorization: `Bearer ${token}`,
    };
  }

  const apolloOptions = {
    connectToDevTools: process.browser && !IS_PRODUCTION,
    ssrMode: !process.browser,
    cache: new InMemoryCache({ dataIdFromObject }).restore(initialState || {}),
    link: ApolloLink.from([new ErrorLink(errorHook), new HttpLink(linkOptions)]),
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
    case "Key":
      return `${object.keyID}`;
    case "NewKey":
      return `${object.keyString}`;
    case "Project":
      return `${object.projectID}`;
    case "Stream":
      return `${object.streamID}`;
    case "Record":
      return `${object.recordID}`;
    default: {
      console.warn(`Unknown typename in dataIdFromObject: ${object.__typename}`);
      return defaultDataIdFromObject(object);
    }
  }
};

const errorHook = ({ graphQLErrors, networkError }) => {
  // redirect to /auth/logout if error is `UNAUTHENTICATED` and `token` is set
  // (probably means the user logged out in another window)
  if (graphQLErrors && graphQLErrors.length > 0) {
    let error = graphQLErrors[0];
    if (error.extensions && error.extensions.code === "UNAUTHENTICATED") {
      if (token) {
        if (process.browser) {
          document.location.href = "/auth/logout";
        } else {
          res.redirect("/auth/logout");
        }
      }
    }
    if (error.extensions && error.extensions.code === "VALIDATION_ERROR") {
      console.log("Validation error", error.extensions.exception);
    }
  }
};
