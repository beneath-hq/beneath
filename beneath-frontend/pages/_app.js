import App, { Container } from "next/app";
import Router from "next/router";
import withGA from "next-ga";
import React from "react";
import withUser from "../hocs/withUser";
import withApollo from "../hocs/withApollo";
import { AuthProvider } from "../hocs/auth";
import { ApolloProvider } from "react-apollo";

class BeneathApp extends App {
  render() {
    const { Component, pageProps, apollo, user } = this.props;
    return (
      <Container>
        <ApolloProvider client={apollo}>
          <AuthProvider user={user}>
            <Component {...pageProps} />
          </AuthProvider>
        </ApolloProvider>
      </Container>
    );
  }
}

export default withApollo(withUser(withGA("UA-118362426-2", Router)(BeneathApp))
);
