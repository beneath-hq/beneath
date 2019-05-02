import App, { Container } from "next/app";
import Router from "next/router";
import withGA from "next-ga";
import React from "react";
import { ThemeProvider } from "@material-ui/styles";
import CssBaseline from "@material-ui/core/CssBaseline";
import theme from "../lib/theme";
import { withApolloClient } from "../hocs/apollo";
import { AuthProvider, withUser } from "../hocs/auth";
import { ApolloProvider } from "react-apollo";

class BeneathApp extends App {
  constructor() {
    super();
  }

  componentDidMount() {
    // For MUI + Next.js, see https://github.com/mui-org/material-ui/blob/master/examples/nextjs/pages/_app.js
    const jssStyles = document.querySelector('#jss-server-side');
    if (jssStyles && jssStyles.parentNode) {
      jssStyles.parentNode.removeChild(jssStyles);
    }
  }

  render() {
    const { Component, apolloClient, user, pageProps } = this.props;

    return (
      <Container>
        <AuthProvider user={user}>
          <ApolloProvider client={apolloClient}>
            <ThemeProvider theme={theme}>
              <CssBaseline />
              <Component {...pageProps} />
            </ThemeProvider>
          </ApolloProvider>
        </AuthProvider>
      </Container>
    );
  }
}

export default withUser(withApolloClient(withGA("UA-118362426-2", Router)(BeneathApp)));
