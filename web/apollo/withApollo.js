import React from "react";
import Head from "next/head";
import { ApolloProvider } from "@apollo/client";

import { getApolloClient } from "./client";

export function withApollo(PageComponent, { ssr = true } = {}) {
  const WithApollo = ({ apolloClient, initialState, ...pageProps }) => {
    const client = apolloClient || getApolloClient({ initialState });
    return (
      <ApolloProvider client={client}>
        <PageComponent {...pageProps} />
      </ApolloProvider>
    );
  };

  // Set the correct displayName in development
  if (process.env.NODE_ENV !== "production") {
    const displayName = PageComponent.displayName || PageComponent.name || "Component";
    WithApollo.displayName = `withApollo(${displayName})`;
    if (displayName === "App") {
      console.warn("This withApollo HOC only works with PageComponents.");
    }
  }

  if (!ssr && !PageComponent.getInitialProps) {
    return WithApollo;
  }

  WithApollo.getInitialProps = async (ctx) => {
    const { AppTree, req, res } = ctx;

    // Initialize ApolloClient, add it to the ctx object so
    // we can use it in `PageComponent.getInitialProp`.
    const apolloClient = getApolloClient({ req, res });
    ctx.apolloClient = apolloClient;

    // Run wrapped getInitialProps methods
    let pageProps = {};
    if (PageComponent.getInitialProps) {
      pageProps = await PageComponent.getInitialProps(ctx);
    }

    // Only on the server:
    if (typeof window === "undefined") {
      // When redirecting, the response is finished.
      // No point in continuing to render
      if (ctx.res && ctx.res.finished) {
        return pageProps;
      }

      // Only if ssr is enabled
      if (ssr) {
        try {
          // Run all GraphQL queries
          const { getDataFromTree } = await import("@apollo/client/react/ssr");
          await getDataFromTree(
            <AppTree
              pageProps={{
                ...pageProps,
                apolloClient,
              }}
            />
          );
        } catch (error) {
          // Prevent Apollo Client GraphQL errors from crashing SSR.
          // Handle them in components via the data.error prop:
          // https://www.apollographql.com/docs/react/api/react-apollo.html#graphql-query-data-error
          // console.error("Error while running `getDataFromTree`", error);
        }

        // getDataFromTree does not call componentWillUnmount
        // head side effect therefore need to be cleared manually
        Head.rewind();
      }
    }

    // Extract query data from the Apollo store
    const initialState = apolloClient.cache.extract();

    return {
      ...pageProps,
      initialState,
    };
  };

  return WithApollo;
}
