import { Typography } from "@material-ui/core";
import { NextPage } from "next";
import React from "react";

import { withApollo } from "apollo/withApollo";
import Page from "components/Page";
import { Paper } from "components/Paper";

const TicketSuccessPage: NextPage = () => {
  return (
    <Page title="Authorize client" contentMarginTop="normal" maxWidth="md">
      <Paper padding="large">
        <Typography variant="h1" component="h2" gutterBottom>
          Authorize client
        </Typography>
        <Typography variant="body1" gutterBottom>
          Success! You can now close this window and continue.
        </Typography>
      </Paper>
    </Page>
  );
};

export default withApollo(TicketSuccessPage);
