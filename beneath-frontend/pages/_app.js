import App, { Container } from "next/app";
import Router from "next/router";
import withGA from "next-ga";
import React from "react";
import { MuiThemeProvider } from "@material-ui/core/styles";
import CssBaseline from "@material-ui/core/CssBaseline";
import JssProvider from "react-jss/lib/JssProvider";
import getPageContext from "../lib/getPageContext";
import { withApolloClient } from "../hocs/apollo";
import { AuthProvider, withUser } from "../hocs/auth";
import { ApolloProvider } from "react-apollo";

class BeneathApp extends App {
  constructor() {
    super();
    this.pageContext = getPageContext();
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
    const { generateClassName, sheetsManager, sheetsRegistry, theme } = this.pageContext;
    return (
      <Container>
        <AuthProvider user={user}>
          <ApolloProvider client={apolloClient}>
            <JssProvider registry={sheetsRegistry} generateClassName={generateClassName}>
              <MuiThemeProvider theme={theme} sheetsManager={sheetsManager}>
                <CssBaseline />
                <Component pageContext={this.pageContext} {...pageProps} />
              </MuiThemeProvider>
            </JssProvider>
          </ApolloProvider>
        </AuthProvider>
      </Container>
    );
  }
}

export default withUser(withApolloClient(withGA("UA-118362426-2", Router)(BeneathApp)));
