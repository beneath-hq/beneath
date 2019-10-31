import { Container, Typography, withStyles } from "@material-ui/core";
import { NextPage } from "next";
import React from "react";

import Page from "../components/Page";

export interface ErrorPageProps {
  apolloError?: any;
  message?: string;
  statusCode?: number;
}

const ErrorPage: NextPage<ErrorPageProps> = ({ apolloError, message, statusCode }) => {
  let title = "An unknown error occurred";
  let details;

  if (statusCode === 404) {
    title = "Sorry, but we couldn't find that page";
  }

  if (statusCode === 401) {
    title = "Please sign in or create a user to view this page";
  }

  if (statusCode) {
    title = `An error with status ${statusCode} occurred`;
  }

  if (message) {
    if (statusCode) {
      details = message;
    } else {
      title = message;
    }
  }

  if (apolloError) {
    if (apolloError.message) {
      title = apolloError.message.replace("GraphQL error: ", "");
    }
  }

  return (
    <Page title={title} contentMarginTop="normal">
      <div>
        <Container maxWidth="lg">
          <Typography component="h2" variant="h4" align="center" gutterBottom>
            Error: {title}
          </Typography>
        </Container>
      </div>
    </Page>
  );
};

ErrorPage.getInitialProps = async ({ res, err }) => {
  const statusCode = res ? res.statusCode : err ? err.statusCode : undefined;
  return { statusCode };
};

export default ErrorPage;
