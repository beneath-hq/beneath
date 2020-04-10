import React, { FC } from "react";
import { useQuery } from "@apollo/react-hooks";
import useMe from "../../hooks/useMe";
import clsx from "clsx";

import NextMuiLinkList from "../NextMuiLinkList";
import ProfileHero from "../ProfileHero";
import ErrorPage from "../ErrorPage";
import Loading from "../Loading";
import Page from "../Page";
import UsageIndicator from "../metrics/user/UsageIndicator";
import ViewUserProjects from "../organization/personal/ViewUserProjects";
import { monthFloor, normalizeMetrics, now, weekAgo, yearAgo } from "../metrics/util";

import { QUERY_USER_BY_USERNAME } from "../../apollo/queries/user";
import { UserByUsername, UserByUsernameVariables } from "../../apollo/types/UserByUsername";
import { GET_USER_METRICS } from "../../apollo/queries/metrics";
import { GetUserMetrics, GetUserMetricsVariables } from "../../apollo/types/GetUserMetrics";

import { Button, Grid, makeStyles, Theme, Typography } from "@material-ui/core";
import TopProjects from "./TopProjects";

const useStyles = makeStyles((theme: Theme) => ({
  buttons: {
    marginTop: theme.spacing(3),
    marginBottom: theme.spacing(6),
  },
  sectionHeader: {
    fontSize: theme.typography.pxToRem(24),
    marginTop: theme.spacing(6),
    marginBottom: theme.spacing(1),
  },
  button: {},
  primaryButton: {},
  secondaryButton: {},
}));

const Springboard: FC = () => {
  const classes = useStyles();

  // GET PROFILE HERO
  const me = useMe();
  if (!me) {
    return <p>Need to log in to view your dashboard -- this shouldn't ever get hit</p>
  }
  const username = me.user.username

  const { loading, error, data } = useQuery<UserByUsername, UserByUsernameVariables>(QUERY_USER_BY_USERNAME, {
    fetchPolicy: "cache-and-network",
    variables: { username },
  });

  // GET METRICS
  const from = monthFloor(yearAgo());
  const until = monthFloor(now());

  const { loading: loading2, error: error2, data: data2 } = useQuery<GetUserMetrics, GetUserMetricsVariables>(GET_USER_METRICS, {
    variables: {
      from: from.toISOString(),
      period: "M",
      userID: me.userID,
    },
  });

  const { metrics, total, latest } = normalizeMetrics(from, until, "month", data2 ? data2.getUserMetrics : null);

  if (loading || loading2) {
    return (
      <Page>
        <Loading justify="center" />
      </Page>
    );
  }

  if (error || error2 || !data) {
    return <ErrorPage apolloError={error} />;
  }

  const user = data.userByUsername;
  if (!user) {
    return <ErrorPage statusCode={404} />;
  }

  return (
    <React.Fragment>
      <ProfileHero name={user.name} description={user.bio} avatarURL={user.photoURL} />
      <Grid container justify="center" spacing={2} item xs={12}>
        <UsageIndicator standalone={true} kind="read" usage={latest.readBytes} quota={me.readQuota} />
        <UsageIndicator standalone={true} kind="write" usage={latest.writeBytes} quota={me.writeQuota} />
      </Grid>
      
      <Typography className={classes.sectionHeader} variant="h3" gutterBottom align="center">My projects</Typography>
      <ViewUserProjects user={user} />

      <Grid className={classes.buttons} container spacing={2} justify="center">
        {/* <Grid item>
          <Button
            size="medium"
            color="default"
            variant="outlined"
            className={clsx(classes.button, classes.secondaryButton)}
            href={`//about.beneath.dev/`}
          >
            Overview
          </Button>
        </Grid> */}
        <Grid item>
          <Button
            size="medium"
            color="default"
            variant="outlined"
            className={clsx(classes.button, classes.secondaryButton)}
            href={`//about.beneath.dev/docs/write-data-from-your-app/`}
            component={NextMuiLinkList}
          >
            Tutorials
          </Button>
        </Grid>
        <Grid item>
          <Button
            size="medium"
            color="default"
            variant="outlined"
            className={clsx(classes.button, classes.secondaryButton)}
            href={`/user?name=${username}&tab=monitoring`}
            component={NextMuiLinkList}
          >
            Monitor
          </Button>
        </Grid>
      </Grid>
      <TopProjects/>
    </React.Fragment>
  )
}

export default Springboard;