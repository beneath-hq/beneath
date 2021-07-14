import { Button, Grid, makeStyles, Theme, Typography } from "@material-ui/core";
import { NextPage } from "next";
import React, { useEffect } from "react";

import { withApollo } from "apollo/withApollo";
import Page from "components/Page";
import { Paper } from "components/Paper";
import { useRouter } from "next/router";
import { AuthTicketByID, AuthTicketByIDVariables } from "apollo/types/AuthTicketByID";
import { QUERY_AUTH_TICKET, UPDATE_AUTH_TICKET } from "apollo/queries/user";
import { useMutation, useQuery } from "@apollo/client";
import ErrorPage from "components/ErrorPage";
import Loading from "components/Loading";
import useMe from "hooks/useMe";
import { setRedirectAfterAuth } from "lib/authRedirect";
import { UpdateAuthTicket, UpdateAuthTicketVariables } from "apollo/types/UpdateAuthTicket";

const useStyles = makeStyles((theme: Theme) => ({
  buttons: {
    marginTop: theme.spacing(2.5),
  },
}));

const TicketPage: NextPage = () => {
  const me = useMe();
  const router = useRouter();
  const classes = useStyles();

  const ticketID = typeof router.query.id === "string" ? router.query.id : "";

  useEffect(() => {
    if (typeof window !== "undefined") {
      if (ticketID === "") {
        router.replace("/");
      }

      if (!me) {
        setRedirectAfterAuth(router.pathname, router.query, router.asPath);
        router.replace("/-/auth");
      }
    }
  }, []);

  const { loading, error, data } = useQuery<AuthTicketByID, AuthTicketByIDVariables>(QUERY_AUTH_TICKET, {
    variables: { authTicketID: ticketID },
    skip: ticketID === "",
  });

  const [updateAuthTicket, { error: updateError, loading: updateLoading }] = useMutation<
    UpdateAuthTicket,
    UpdateAuthTicketVariables
  >(UPDATE_AUTH_TICKET, {
    onCompleted: (data) => {
      if (data.updateAuthTicket) {
        router.replace("/-/auth/ticket/success");
      } else {
        router.replace("/-/auth/ticket/denied");
      }
    },
  });

  if (error) {
    return <ErrorPage apolloError={error} />;
  }

  if (updateError) {
    return <ErrorPage apolloError={updateError} />;
  }

  if (loading || !data || !me || !ticketID) {
    return (
      <Page title="Authorize client" contentMarginTop="normal">
        <Loading justify="center" />
      </Page>
    );
  }

  const ticket = data.authTicketByID;
  return (
    <Page title="Authorize client" contentMarginTop="normal" maxWidth="md">
      <Paper padding="large">
        <Typography variant="h1" component="h2" gutterBottom>
          Authorize client
        </Typography>
        <Typography variant="body1" gutterBottom>
          <strong>{ticket.requesterName}</strong> is requesting permission to access Beneath on your behalf. It will get
          full access to your account. You can revoke access at any time from your secrets page.
        </Typography>
        <Typography variant="body1" gutterBottom>
          If you did not explicitly begin this request, then deny it.
        </Typography>
        <Grid className={classes.buttons} container spacing={2} direction="row">
          <Grid item>
            <Button
              variant="contained"
              color="primary"
              disabled={updateLoading}
              onClick={() =>
                updateAuthTicket({ variables: { input: { authTicketID: ticket.authTicketID, approve: true } } })
              }
            >
              Approve
            </Button>
          </Grid>
          <Grid item>
            <Button
              variant="contained"
              disabled={updateLoading}
              onClick={() =>
                updateAuthTicket({ variables: { input: { authTicketID: ticket.authTicketID, approve: false } } })
              }
            >
              Deny
            </Button>
          </Grid>
        </Grid>
      </Paper>
    </Page>
  );
};

export default withApollo(TicketPage);
