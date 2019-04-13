import withApollo from "next-with-apollo";
import { ApolloClient } from "apollo-client";
import { InMemoryCache } from "apollo-cache-inmemory";
import { HttpLink } from "apollo-link-http";
import { onError } from "apollo-link-error";
import { ApolloLink } from "apollo-link";
import { HTTP_PROTOCOL, API_HOST, WEBSOCKET_PROTOCOL } from "../lib/connection";
import Cookies from 'js-cookie';

const authMiddleware = new ApolloLink((operation, forward) => {
  const token = Cookies.get("idToken") || null;
  // add the authorization to the headers
  operation.setContext({
    headers: {
      authorization: `Bearer ${token}`
    }
  })
  return forward(operation)
})

const compositeLink = ApolloLink.from([
  authMiddleware,
  onError(({ graphQLErrors, networkError }) => {
    if (graphQLErrors)
      graphQLErrors.map(({ message, locations, path }) =>
        console.log(
          `[GraphQL error]: Message: ${message}, Location: ${locations}, Path: ${path}`
        )
      );
    if (networkError) console.log(`[Network error]: ${networkError}`);
  }),
  new HttpLink({
    uri: `${HTTP_PROTOCOL}://${API_HOST}/graphql`,
    credentials: "same-origin",
  }),
]);

export default withApollo(
  ({ ctx, headers, initialState }) => {
    return new ApolloClient({
      link: compositeLink,
      cache: new InMemoryCache().restore(initialState || {}),
    })
  }
);
