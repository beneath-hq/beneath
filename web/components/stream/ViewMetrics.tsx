import { FC } from "react";

import { Grid } from "@material-ui/core";

import { EntityKind } from "../../apollo/types/globalTypes";
import { StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName } from "../../apollo/types/StreamByOrganizationProjectAndName";
import ErrorNote from "../ErrorNote";
import { useMonthlyMetrics, useWeeklyMetrics } from "../metrics/hooks";
import BatchTopIndicators from "../metrics/stream/BatchTopIndicators";
import StreamingTopIndicators from "../metrics/stream/StreamingTopIndicators";
import WeekChart from "../metrics/WeekChart";

export interface ViewMetricsProps {
  stream: StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName;
}

const ViewMetrics: FC<ViewMetricsProps> = ({ stream }) => {
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

export default ViewMetrics;

const StreamingMetricsOverview: FC<ViewMetricsProps> = ({ stream }) => {
  const { metrics, total, latest, error } = useMonthlyMetrics(EntityKind.Stream, stream.streamID);
  return (
    <>
      <StreamingTopIndicators latest={latest} total={total} period="month" totalPeriod="all time" />
      {error && <ErrorNote error={error} />}
    </>
  );
};

const StreamingMetricsWeek: FC<ViewMetricsProps> = ({ stream }) => {
  const { metrics, total, latest, error } = useWeeklyMetrics(EntityKind.Stream, stream.streamID);
  return (
    <>
      <WeekChart metrics={metrics} y1="writeRecords" title={"Rows written in the last 7 days"} />
      {error && <ErrorNote error={error} />}
    </>
  );
};

const BatchMetricsOverview: FC<ViewMetricsProps> = ({ stream }) => {
  const { metrics, total, latest, error } = useMonthlyMetrics(EntityKind.Stream, stream.streamID);
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

const BatchMetricsWeek: FC<ViewMetricsProps> = ({ stream }) => {
  const { metrics, total, latest, error } = useWeeklyMetrics(EntityKind.Stream, stream.streamID);
  return (
    <>
      <WeekChart metrics={metrics} y1="readRecords" title={"Rows read in the last 7 days"} />
      {error && <ErrorNote error={error} />}
    </>
  );
};
