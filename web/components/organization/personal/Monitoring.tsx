import { useQuery } from "@apollo/react-hooks";
import { FC } from "react";

import { GET_USER_METRICS } from "../../../apollo/queries/metrics";
import { GetUserMetrics, GetUserMetricsVariables } from "../../../apollo/types/GetUserMetrics";
import { Me_me } from "../../../apollo/types/Me";
import TopIndicators from "../../metrics/user/TopIndicators";
import UsageIndicator from "../../metrics/user/UsageIndicator";
import { hourFloor, monthFloor, normalizeMetrics, now, weekAgo, yearAgo } from "../../metrics/util";
import WeekChart from "../../metrics/WeekChart";

import { Grid, makeStyles, Theme, Typography } from "@material-ui/core";

const useStyles = makeStyles((theme: Theme) => ({
}));

interface MeProps {
  me: Me_me;
}

const UserMetrics: FC<MeProps> = ({ me }) => {
  return (
    <Grid container spacing={2}>
      <UserMetricsOverview me={me} />
      <UserMetricsWeek me={me} />
    </Grid>
  );
};

export default UserMetrics;

const UserMetricsOverview: FC<MeProps> = ({ me }) => {
  const from = monthFloor(yearAgo());
  const until = monthFloor(now());

  const { loading, error, data } = useQuery<GetUserMetrics, GetUserMetricsVariables>(GET_USER_METRICS, {
    variables: {
      from: from.toISOString(),
      period: "M",
      userID: me.userID,
    },
  });

  const { metrics, total, latest } = normalizeMetrics(from, until, "month", data ? data.getUserMetrics : null);

  return (
    <>
      <Grid container spacing={2} item xs={12}>
        <UsageIndicator standalone={true} kind="read" usage={latest.readBytes} quota={me.readQuota} />
        <UsageIndicator standalone={true} kind="write" usage={latest.writeBytes} quota={me.writeQuota} />
      </Grid>
      <TopIndicators latest={latest} total={total} />
      {error && <ErrorNote error={error} />}
    </>
  );
};

const UserMetricsWeek: FC<MeProps> = ({ me }) => {
  const from = hourFloor(weekAgo());
  const until = hourFloor(now());

  const { loading, error, data } = useQuery<GetUserMetrics, GetUserMetricsVariables>(GET_USER_METRICS, {
    variables: {
      from: from.toISOString(),
      period: "H",
      userID: me.userID,
    },
  });

  const { metrics, total, latest } = normalizeMetrics(from, until, "hour", data ? data.getUserMetrics : null);

  return (
    <>
      <WeekChart metrics={metrics} y1={"readBytes"} title="Data read in the last 7 days" />
      {error && <ErrorNote error={error} />}
    </>
  );
};

const ErrorNote: FC<{ error: Error }> = ({ error }) => {
  return (
    <Grid item xs={12}>
      <Typography color="error">{error.message}</Typography>
    </Grid>
  );
};
