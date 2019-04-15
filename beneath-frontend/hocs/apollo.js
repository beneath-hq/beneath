// Polyfill fetch() on the server (used by apollo-client)
import fetch from 'isomorphic-unfetch'
if (!process.browser) {
  global.fetch = fetch
}

import { ApolloClient, InMemoryCache, HttpLink } from 'apollo-boost'
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
}

function createApolloClient({ initialState, token }) {
  let linkOptions = {
    uri: `${API_URL}/graphql`,
    credentials: "include",
  };

  if (token) {
    linkOptions.headers = {
      "Authorization": `Bearer ${token}`,
    };
  }

  let apolloOptions = {
    connectToDevTools: process.browser && !IS_PRODUCTION,
    ssrMode: !process.browser,
    cache: new InMemoryCache().restore(initialState || {}),
    link: new HttpLink(linkOptions),
  };

  return new ApolloClient(apolloOptions);
}

export const withApolloClient = (App) => {
  return class Apollo extends React.Component {
    static displayName = 'withApollo(App)'
    static async getInitialProps(ctx) {
      const { Component, router, user } = ctx

      // Get app props to pass on
      let appProps = {}
      if (App.getInitialProps) {
        appProps = await App.getInitialProps({ ...ctx })
      }

      // Get token from user
      let token = null;
      if (user) {
        token = user.token;
      }
      
      // Get apollo client
      const apollo = getApolloClient({ token })
      
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
}
