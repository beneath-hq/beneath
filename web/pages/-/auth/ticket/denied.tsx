import { Typography } from "@material-ui/core";
import { NextPage } from "next";
import React from "react";

import { withApollo } from "apollo/withApollo";
import Page from "components/Page";
import { Paper } from "components/Paper";
import { Link } from "components/Link";

const TicketDeniedPage: NextPage = () => {
  return (
    <Page title="Authorize client" contentMarginTop="normal" maxWidth="md">
      <Paper padding="large">
        <Typography variant="h1" component="h2" gutterBottom>
          Authorize client
        </Typography>
        <Typography variant="body1" gutterBottom>
          You have denied the request. If you suspect malicious activity, please{" "}
          <Link href="https://about.beneath.dev/contact/">report it</Link>.
        </Typography>
      </Paper>
    </Page>
  );
};

export default withApollo(TicketDeniedPage);
