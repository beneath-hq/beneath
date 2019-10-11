import App from "next/app";
import Router from "next/router";
import React from "react";
import { ThemeProvider } from "@material-ui/styles";
import CssBaseline from "@material-ui/core/CssBaseline";
import theme from "../lib/theme";
import { withApolloClient } from "../apollo/withApollo";
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
    const { Component, apolloClient, token, pageProps } = this.props;

    return (
      <ApolloProvider client={apolloClient}>
        <ThemeProvider theme={theme}>
          <CssBaseline />
          <Component {...pageProps} />
        </ThemeProvider>
      </ApolloProvider>
    );
  }
}

export default withApolloClient(BeneathApp);
