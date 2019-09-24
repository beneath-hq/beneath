import Head from "next/head";
import React from "react";
import { getDataFromTree } from "react-apollo";

import { getApolloClient } from "./client";

export const withApolloClient = (App) => {
  return class Apollo extends React.Component {
    static displayName = "withApollo(App)";

    constructor(props) {
      super(props);
      this.apolloClient = getApolloClient({
        token: props.token,
        initialState: props.apolloState,
      });
    }

    static async getInitialProps(ctx) {
      const {
        Component,
        router,
        ctx: { req, res },
      } = ctx;

      // read token from token
      let token = readTokenFromCookie(req);

      // Get app props to pass on
      let appProps = {};
      if (App.getInitialProps) {
        appProps = await App.getInitialProps({ ...ctx });
      }

      // Get apollo client
      const apollo = getApolloClient({ token, res });

      // Run all GraphQL queries in the component tree
      // and extract the resulting data
      if (!process.browser) {
        try {
          // Run all GraphQL queries
          await getDataFromTree(
            <App {...appProps} Component={Component} router={router} apolloClient={apollo} token={token} />
          );
        } catch (error) {
          // Prevent Apollo Client GraphQL errors from crashing SSR.
          // Handle them in components via the data.error prop:
          // https://www.apollographql.com/docs/react/api/react-apollo.html#graphql-query-data-error
          console.error("Error while running `getDataFromTree`", error);
        }

        // getDataFromTree does not call componentWillUnmount
        // head side effect therefore need to be cleared manually
        Head.rewind();
      }

      // Extract query data from the Apollo store
      const apolloState = apollo.cache.extract();

      return {
        ...appProps,
        apolloState,
        token,
      };
    }

    render() {
      return <App {...this.props} apolloClient={this.apolloClient} />;
    }
  };
};

const readTokenFromCookie = (maybeReq) => {
  let token = null;
  let cookie = null;
  if (maybeReq) {
    cookie = maybeReq.headers.cookie;
  } else if (document) {
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
