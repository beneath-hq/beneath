import Head from "next/head";
import React from "react";
import { getDataFromTree } from "react-apollo";
import { getApollo } from "../lib/apollo";

import { getUser } from "../lib/auth";

const makeJWTGetter = ctx => {
  return () => {
    let user = getUser(ctx ? ctx.req : null);
    if (user) {
      return user.jwt;
    }
    return null;
  };
};

export default App => {
  return class Apollo extends React.Component {
    static displayName = "withApollo(App)";
    static async getInitialProps({ Component, router, ctx }) {
      let appProps = {};
      if (App.getInitialProps) {
        appProps = await App.getInitialProps({ Component, router, ctx });
      }

      // Run all GraphQL queries in the component tree
      // and extract the resulting data
      const apollo = getApollo(null, makeJWTGetter(ctx));
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
        apolloClient: apollo,
      };
    }

    constructor(props) {
      super(props);
      if (props.apolloClient) {
        this.apolloClient = props.apolloClient;
      } else {
        this.apolloClient = getApollo(props.apolloState, makeJWTGetter());
      }
    }

    render() {
      return <App {...this.props} apolloClient={this.apolloClient} />;
    }
  };
};
