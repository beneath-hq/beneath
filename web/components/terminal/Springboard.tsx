import { useQuery } from "@apollo/react-hooks";
import clsx from "clsx";
import React, { FC } from "react";

import { Button, Grid, makeStyles, Theme, Typography } from "@material-ui/core";

import useMe from "../../hooks/useMe";
import { toURLName } from "../../lib/names";
import Loading from "../Loading";
import UsageIndicator from "../metrics/user/UsageIndicator";
import { monthFloor, normalizeMetrics, now, yearAgo } from "../metrics/util";
import NextMuiLinkList from "../NextMuiLinkList";
import ViewProjects from "../organization/ViewProjects";
import ProfileHero from "../ProfileHero";
import TopProjects from "./TopProjects";

import { GET_USER_METRICS } from "../../apollo/queries/metrics";
import { GetUserMetrics, GetUserMetricsVariables } from "../../apollo/types/GetUserMetrics";

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

  const me = useMe();
  if (!me || !me.personalUserID) {
    return <p>Need to log in to view your dashboard -- this shouldn't ever get hit</p>
  }

  // GET METRICS
  const from = monthFloor(yearAgo());
  const until = monthFloor(now());

  const { loading, error, data } = useQuery<GetUserMetrics, GetUserMetricsVariables>(GET_USER_METRICS, {
    variables: {
      from: from.toISOString(),
      period: "M",
      userID: me.personalUserID,
    },
  });

  if (loading) {
    return <Loading justify="center" />;
  }

  if (error || !data) {
    return <p>Error: {JSON.stringify(error)}</p>;
  }

  const { metrics, total, latest } = normalizeMetrics(from, until, "month", data.getUserMetrics);

  return (
    <>
      <ProfileHero
        name={toURLName(me.name)}
        displayName={me.displayName}
        description={me.description}
        avatarURL={me.photoURL}
      />
      {/* <Grid container justify="center" spacing={2} item xs={12}>
        <UsageIndicator standalone={true} kind="read" usage={latest.readBytes} quota={me.readQuota} />
        <UsageIndicator standalone={true} kind="write" usage={latest.writeBytes} quota={me.writeQuota} />
      </Grid> */}

      <Typography className={classes.sectionHeader} variant="h3" gutterBottom align="center">
        My projects
      </Typography>
      <ViewProjects organization={me} />

      <Grid className={classes.buttons} container spacing={2} justify="center">
        <Grid item>
          <Button
            size="medium"
            color="default"
            variant="outlined"
            className={clsx(classes.button, classes.secondaryButton)}
            href={`https://about.beneath.dev/docs/quick-starts/write-data-from-your-app/`}
            component={NextMuiLinkList}
          >
            Create Project
          </Button>
        </Grid>
        <Grid item>
          <Button
            size="medium"
            color="default"
            variant="outlined"
            className={clsx(classes.button, classes.secondaryButton)}
            href={`/organization?organization_name=${toURLName(me.name)}&tab=monitoring`}
            as={`/${toURLName(me.name)}/-/monitoring`}
            component={NextMuiLinkList}
          >
            Monitor
          </Button>
        </Grid>
      </Grid>
      <TopProjects />
    </>
  );
}

export default Springboard;
