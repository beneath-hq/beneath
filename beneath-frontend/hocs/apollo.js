// Polyfill fetch() on the server (used by apollo-client)
import fetch from 'isomorphic-unfetch'
if (!process.browser) {
  global.fetch = fetch
}

import { ApolloClient, ApolloLink, HttpLink, InMemoryCache, defaultDataIdFromObject } from 'apollo-boost'
import { ErrorLink } from 'apollo-link-error'
import Head from 'next/head'
import React from 'react'
import { getDataFromTree } from 'react-apollo'

import { API_URL, IS_PRODUCTION } from "../lib/connection";

let apolloClient = null;

const getApolloClient = (options) => {
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
      "Authorization": `Bearer ${token}`,
    };
  }

  // returns ID for objects for caching/automatic state update
  const dataIdFromObject = (object) => {
    switch (object.__typename) {
      case "User": return object.userId;
      case "Me": return `me:${object.userId}`;
      case "Key": return `${object.keyId}`;
      case "NewKey": return `${object.keyString}`;
      case "Project": return `${object.projectId}`;
      default: {
        console.warn(`Unknown typename in dataIdFromObject: ${object.__typename}`);
        return defaultDataIdFromObject(object);
      };
    }
  };

  const errorHook = ({graphQLErrors, networkError}) => {
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

  const apolloOptions = {
    connectToDevTools: process.browser && !IS_PRODUCTION,
    ssrMode: !process.browser,
    cache: new InMemoryCache({ dataIdFromObject }).restore(initialState || {}),
    link: ApolloLink.from([new ErrorLink(errorHook), new HttpLink(linkOptions)]),
  };

  return new ApolloClient(apolloOptions);
};

export const withApolloClient = (App) => {
  return class Apollo extends React.Component {
    static displayName = 'withApollo(App)'
    static async getInitialProps(ctx) {
      const { Component, router, token, ctx: { res } } = ctx

      // Get app props to pass on
      let appProps = {}
      if (App.getInitialProps) {
        appProps = await App.getInitialProps({ ...ctx })
      }
      
      // Get apollo client
      const apollo = getApolloClient({ token, res })
      
      // Run all GraphQL queries in the component tree
      // and extract the resulting data
      if (!process.browser) {
        try {
          // Run all GraphQL queries
          await getDataFromTree(
            <App
              {...appProps}
              Component={Component}
              router={router}
              apolloClient={apollo}
              token={token}
            />
          )
        } catch (error) {
          // Prevent Apollo Client GraphQL errors from crashing SSR.
          // Handle them in components via the data.error prop:
          // https://www.apollographql.com/docs/react/api/react-apollo.html#graphql-query-data-error
          console.error('Error while running `getDataFromTree`', error)
        }

        // getDataFromTree does not call componentWillUnmount
        // head side effect therefore need to be cleared manually
        Head.rewind()
      }

      // Extract query data from the Apollo store
      const apolloState = apollo.cache.extract()

      return {
        ...appProps,
        apolloState,
        token,
      }
    }

    constructor(props) {
      super(props)
      this.apolloClient = getApolloClient({
        token: props.token,
        initialState: props.apolloState,
      });
    }

    render() {
      return <App {...this.props} apolloClient={this.apolloClient} />
    }
  }
};
