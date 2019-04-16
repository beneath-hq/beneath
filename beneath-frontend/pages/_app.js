import App, { Container } from "next/app";
import Router from "next/router";
import withGA from "next-ga";
import React from "react";
import { withApolloClient } from "../hocs/apollo";
import { AuthProvider, withUser } from "../hocs/auth";
import { ApolloProvider } from "react-apollo";

class BeneathApp extends App {
  render() {
    const { Component, apolloClient, user, pageProps } = this.props;
    return (
      <Container>
        <AuthProvider user={user}>
          <ApolloProvider client={apolloClient}>
            <Component {...pageProps} />
          </ApolloProvider>
        </AuthProvider>
      </Container>
    );
  }
}

export default withUser(withApolloClient(withGA("UA-118362426-2", Router)(BeneathApp)));
