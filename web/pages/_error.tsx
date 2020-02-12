import { Container, Typography, withStyles } from "@material-ui/core";
import { NextPage } from "next";
import React from "react";

import { withApollo } from "../apollo/withApollo";
import ErrorPage from "../components/ErrorPage";

export interface ErrorPageProps {
  statusCode?: number;
}

const errorPage: NextPage<ErrorPageProps> = ({ statusCode }) => {
  return <ErrorPage statusCode={statusCode} />;
};

errorPage.getInitialProps = async ({ res, err }) => {
  const statusCode = res ? res.statusCode : err ? err.statusCode : undefined;
  return { statusCode };
};

export default withApollo(errorPage);
