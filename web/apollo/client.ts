import { ApolloClient, ApolloLink, HttpLink, InMemoryCache, NormalizedCacheObject } from "@apollo/client";
import { onError } from "@apollo/client/link/error";
import { IncomingMessage, ServerResponse } from "http";

import fetch from "isomorphic-unfetch";

import { resolvers, typeDefs } from "apollo/schema";
import { GET_AID, GET_TOKEN } from "apollo/queries/local/token";
import { API_URL, IS_EE, IS_PRODUCTION } from "lib/connection";

import possibleTypesCE from "apollo/possibleTypes.json";
import possibleTypesEE from "ee/apollo/possibleTypes.json";
import typePoliciesCE from "apollo/typePolicies";
import typePoliciesEE from "ee/apollo/typePolicies";

let apolloClient: ApolloClient<NormalizedCacheObject>;

export type GetApolloClientOptions = {
  req?: IncomingMessage;
  res?: ServerResponse;
  initialState?: NormalizedCacheObject;
};

export const getApolloClient = ({ req, res, initialState }: GetApolloClientOptions) => {
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

const createApolloClient = ({ req, res, initialState }: GetApolloClientOptions) => {
  // setup type info for in-memory cache
  const possibleTypes = !IS_EE ? possibleTypesCE : { ...possibleTypesCE, ...possibleTypesEE };
  const typePolicies = !IS_EE ? typePoliciesCE : { ...typePoliciesCE, ...typePoliciesEE };

  // configure cache
  const cache = new InMemoryCache({ possibleTypes, typePolicies }).restore(initialState || {});

  // prepare headers
  const headers = {} as any;

  // read token from cookie
  const token = readTokenFromCookie(req);
  if (token) {
    headers.Authorization = `Bearer ${token}`;
    cache.writeQuery({ query: GET_TOKEN, data: { token } });
  }

  // read anonymous id from cookie
  const anonymousID = readAnonymousIDFromCookie(req);
  if (anonymousID) {
    headers["X-Beneath-Aid"] = anonymousID;
    cache.writeQuery({ query: GET_AID, data: { aid: anonymousID } });
  }

  // make error link
  const errorLink = onError(({ networkError }) => {
    // redirect to /auth/logout if error is `unauthenticated` (probably means the user logged out in another window)
    if (networkError && "result" in networkError) {
      if (networkError.result.error?.match(/authentication error.*/)) {
        if (token) {
          if (process.browser) {
            document.location.href = "/-/redirects/auth/logout";
          } else {
            if (res && !res.headersSent) {
              res.writeHead(307, { Location: "/-/redirects/auth/logout" });
              res.end();
            }
          }
        }
      }
    }
  });

  // http link
  let httpLink: ApolloLink = new HttpLink({
    credentials: "include",
    fetch,
    headers,
    uri: `${API_URL}/graphql`,
  });

  // adds ee link
  if (IS_EE) {
    const eeHttpLink = new HttpLink({
      credentials: "include",
      fetch,
      headers,
      uri: `${API_URL}/ee/graphql`,
    });

    httpLink = ApolloLink.split(
      (operation) => operation.getContext().ee,
      eeHttpLink,
      httpLink,
    );
  }

  return new ApolloClient({
    connectToDevTools: typeof window === "undefined" && !IS_PRODUCTION,
    ssrMode: typeof window === "undefined",
    ssrForceFetchDelay: 100,
    link: ApolloLink.from([errorLink, httpLink]),
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

const readTokenFromCookie = (maybeReq?: IncomingMessage) => {
  let token = null;
  let cookie = null;
  if (maybeReq) {
    cookie = maybeReq.headers.cookie;
  } else if (typeof document !== "undefined") {
    cookie = document.cookie;
  }
  if (cookie) {
    cookie = cookie.split(";").find((c: string) => c.trim().startsWith("token="));
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

const readAnonymousIDFromCookie = (maybeReq?: IncomingMessage) => {
  let aid = null;
  let cookie = null;
  if (maybeReq) {
    cookie = maybeReq.headers.cookie;
  } else if (typeof document !== "undefined") {
    cookie = document.cookie;
  }
  if (cookie) {
    cookie = cookie.split(";").find((c: string) => c.trim().startsWith("aid="));
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
