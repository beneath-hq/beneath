import { FC } from "react";
import { useQuery } from "react-apollo";

import { GET_STREAM_METRICS } from "../../apollo/queries/metrics";
import { GetStreamMetrics, GetStreamMetricsVariables } from "../../apollo/types/GetStreamMetrics";
import { QueryStream } from "../../apollo/types/QueryStream";
import StreamingTopIndicators from "../metrics/stream/StreamingTopIndicators";
import { hourFloor, monthFloor, normalizeMetrics, now, weekAgo, yearAgo } from "../metrics/util";
import WeekChart from "../metrics/WeekChart";

import { Grid, makeStyles, Theme, Typography } from "@material-ui/core";

const useStyles = makeStyles((theme: Theme) => ({
}));

const StreamMetrics: FC<QueryStream> = ({ stream }) => {
  return (
    <Grid container spacing={2}>
      <StreamMetricsOverview stream={stream} />
      <StreamMetricsWeek stream={stream} />
    </Grid>
  );
};

export default StreamMetrics;

const StreamMetricsOverview: FC<QueryStream> = ({ stream }) => {
  const from = monthFloor(yearAgo());
  const until = monthFloor(now());

  const { loading, error, data } = useQuery<GetStreamMetrics, GetStreamMetricsVariables>(GET_STREAM_METRICS, {
    variables: {
      from: from.toISOString(),
      period: "M",
      streamID: stream.streamID,
    },
  });

  const { metrics, total, latest } = normalizeMetrics(from, until, "month", data ? data.getStreamMetrics : null);

  return (
    <>
      <StreamingTopIndicators latest={latest} total={total} period="month" totalPeriod="all time" />
      {error && <ErrorNote error={error} />}
    </>
  );
};

const StreamMetricsWeek: FC<QueryStream> = ({ stream }) => {
  const from = hourFloor(weekAgo());
  const until = hourFloor(now());

  const { loading, error, data } = useQuery<GetStreamMetrics, GetStreamMetricsVariables>(GET_STREAM_METRICS, {
    variables: {
      from: from.toISOString(),
      period: "H",
      streamID: stream.streamID,
    },
  });

  const { metrics, total, latest } = normalizeMetrics(from, until, "hour", data ? data.getStreamMetrics : null);

  return (
    <>
      <WeekChart metrics={metrics} y1="writeRecords" title={"Rows written in the last 7 days"} />
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
