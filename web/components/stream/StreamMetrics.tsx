import { useQuery } from "@apollo/react-hooks";
import { FC } from "react";

import { GET_STREAM_METRICS } from "../../apollo/queries/metrics";
import { GetStreamMetrics, GetStreamMetricsVariables } from "../../apollo/types/GetStreamMetrics";
import { QueryStream_streamByProjectAndName } from "../../apollo/types/QueryStream";
import BatchTopIndicators from "../metrics/stream/BatchTopIndicators";
import StreamingTopIndicators from "../metrics/stream/StreamingTopIndicators";
import { hourFloor, monthFloor, normalizeMetrics, now, weekAgo, yearAgo } from "../metrics/util";
import WeekChart from "../metrics/WeekChart";

import { Grid, makeStyles, Theme, Typography } from "@material-ui/core";

const useStyles = makeStyles((theme: Theme) => ({
  noDataCaption: {
    color: theme.palette.text.secondary,
  },
}));

interface StreamMetricsProps {
  stream: QueryStream_streamByProjectAndName;
}

const StreamMetrics: FC<StreamMetricsProps> = ({ stream }) => {
  const classes = useStyles();

  if (!stream.currentStreamInstanceID) {
    return (
      <Typography className={classes.noDataCaption} variant="body1" align="center">
        There are no metrics to show for this stream because no data has been committed to it
      </Typography>
    );
  }

  if (stream.batch) {
    return (
      <Grid container spacing={2}>
        <BatchMetricsOverview stream={stream} />
        <BatchMetricsWeek stream={stream} />
      </Grid>
    );
  } else {
    return (
      <Grid container spacing={2}>
        <StreamingMetricsOverview stream={stream} />
        <StreamingMetricsWeek stream={stream} />
      </Grid>
    );
  }
};

export default StreamMetrics;

const useMonthlyData = (streamID: string) => {
  const from = monthFloor(yearAgo());
  const until = monthFloor(now());

  const { loading, error, data } = useQuery<GetStreamMetrics, GetStreamMetricsVariables>(GET_STREAM_METRICS, {
    variables: {
      streamID,
      from: from.toISOString(),
      period: "M",
    },
  });

  const { metrics, total, latest } = normalizeMetrics(from, until, "month", data ? data.getStreamMetrics : null);

  return { metrics, total, latest, error, loading };
};

const useWeeklyData = (streamID: string) => {
  const from = hourFloor(weekAgo());
  const until = hourFloor(now());

  const { loading, error, data } = useQuery<GetStreamMetrics, GetStreamMetricsVariables>(GET_STREAM_METRICS, {
    variables: {
      streamID,
      from: from.toISOString(),
      period: "H",
    },
  });

  const { metrics, total, latest } = normalizeMetrics(from, until, "hour", data ? data.getStreamMetrics : null);

  return { metrics, total, latest, error, loading };
};

const StreamingMetricsOverview: FC<StreamMetricsProps> = ({ stream }) => {
  const { metrics, total, latest, error } = useMonthlyData(stream.streamID);
  return (
    <>
      <StreamingTopIndicators latest={latest} total={total} period="month" totalPeriod="all time" />
      {error && <ErrorNote error={error} />}
    </>
  );
};

const StreamingMetricsWeek: FC<StreamMetricsProps> = ({ stream }) => {
  const { metrics, total, latest, error } = useWeeklyData(stream.streamID);
  return (
    <>
      <WeekChart metrics={metrics} y1="writeRecords" title={"Rows written in the last 7 days"} />
      {error && <ErrorNote error={error} />}
    </>
  );
};

const BatchMetricsOverview: FC<StreamMetricsProps> = ({ stream }) => {
  const { metrics, total, latest, error } = useMonthlyData(stream.streamID);
  return (
    <>
      <BatchTopIndicators
        latest={latest}
        total={total}
        period="month"
        instancesCreated={stream.instancesCreatedCount}
        instancesCommitted={stream.instancesCommittedCount}
      />
      {error && <ErrorNote error={error} />}
    </>
  );
};

const BatchMetricsWeek: FC<StreamMetricsProps> = ({ stream }) => {
  const { metrics, total, latest, error } = useWeeklyData(stream.streamID);
  return (
    <>
      <WeekChart metrics={metrics} y1="readRecords" title={"Rows read in the last 7 days"} />
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
