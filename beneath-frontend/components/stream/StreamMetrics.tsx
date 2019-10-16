import numbro from "numbro";
import { FC } from "react";
import { useQuery } from "react-apollo";
import { createClassFromSpec } from "react-vega";

import { GET_STREAM_METRICS } from "../../apollo/queries/metrics";
import { GetStreamMetrics, GetStreamMetricsVariables } from "../../apollo/types/GetStreamMetrics";
import { QueryStream } from "../../apollo/types/QueryStream";

import { Grid, makeStyles, Paper, Theme, Typography } from "@material-ui/core";

interface Metrics {
  readOps: number;
  readBytes: number;
  readRecords: number;
  writeOps: number;
  writeBytes: number;
  writeRecords: number;
}

const blankMetrics = (): Metrics => {
  return {
    readOps: 0,
    readBytes: 0,
    readRecords: 0,
    writeOps: 0,
    writeBytes: 0,
    writeRecords: 0,
  };
};

const computeAggregateMetrics = (metrics: Metrics[] | null): Metrics => {
  const result = blankMetrics();
  if (metrics) {
    for (const m of metrics) {
      result.readOps += m.readOps;
      result.readBytes += m.readBytes;
      result.readRecords += m.readRecords;
      result.writeOps += m.writeOps;
      result.writeBytes += m.writeBytes;
      result.writeRecords += m.writeRecords;
    }
  }
  return result;
};

const getHourlyTo = () => {
  const now = new Date();
  return new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), now.getUTCDate(), now.getUTCHours(), 0, 0, 0));
};

const getHourlyFrom = () => {
  const now = new Date();
  const then = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
  return new Date(Date.UTC(then.getUTCFullYear(), then.getUTCMonth(), then.getUTCDate(), then.getUTCHours(), 0, 0, 0));
};

const getMonthFrom = () => {
  const now = new Date();
  const then = new Date(now.getTime() - 365 * 24 * 60 * 60 * 1000);
  return new Date(Date.UTC(then.getUTCFullYear(), then.getUTCMonth(), 1, 0, 0, 0, 0));
};

const useStyles = makeStyles((theme: Theme) => ({
  numberContainer: {
    padding: "10px",
  },
  vegaChart: {
    width: "100%",
  },
  chartTitle: {
    fontSize: "1rem",
    fontWeight: "bold",
    paddingTop: "5px",
    paddingLeft: "5px",
    paddingBottom: "10px",
    textTransform: "uppercase",
  },
}));

const MetricPaper = ({ top, middle, bottom }: any) => {
  const classes = useStyles();
  return (
    <Paper className={classes.numberContainer}>
    <Typography variant="overline">{top}</Typography>
    <Typography variant="h2" component="p">{middle}</Typography>
    <Typography variant="overline">{bottom}</Typography>
  </Paper>
  );
};

const bytesFormat = { base: "decimal", mantissa: 1, output: "byte" };
const intFormat = { thousandSeparated: true };

interface MetricsOverviewProps {
  current: Metrics;
  total: Metrics;
  alltime: boolean;
}

const MetricsOverview: FC<MetricsOverviewProps> = ({ current, total, alltime }) => {
  const formatTop = (msg: string) => alltime ? `Total ${msg}` : `${msg} This Week`;
  const formatBottom = (msg: string) => alltime ? `(${msg} this month)` : `(${msg} last hour)`;
  return (
    <>
      <Grid item xs={6} md={3}>
        <MetricPaper
          top={formatTop("Records Written")}
          middle={numbro(total.writeRecords).format(intFormat)}
          bottom={formatBottom(numbro(current.writeRecords).format(intFormat))}
        />
      </Grid>
      <Grid item xs={6} md={3}>
        <Paper>
          <MetricPaper
            top={formatTop("Bytes Written")}
            middle={numbro(total.writeBytes).format(bytesFormat)}
            bottom={formatBottom(numbro(current.writeBytes).format(bytesFormat))}
          />
        </Paper>
      </Grid>
      <Grid item xs={6} md={3}>
        <Paper>
          <MetricPaper
            top={formatTop("Read Calls")}
            middle={numbro(total.readOps).format(intFormat)}
            bottom={formatBottom(numbro(current.readOps).format(intFormat))}
          />
        </Paper>
      </Grid>
      <Grid item xs={6} md={3}>
        <Paper>
          <MetricPaper
            top={formatTop("Records Read")}
            middle={numbro(total.readRecords).format(intFormat)}
            bottom={formatBottom(numbro(current.readRecords).format(intFormat))}
          />
        </Paper>
      </Grid>
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

const markColor = "#3caaff";

const MetricsChartVega = createClassFromSpec({
  mode: "vega-lite",
  spec: {
    $schema: "https://vega.github.io/schema/vega-lite/v4.0.0-beta.10.json",
    data: { name: "metrics" },
    mark: { type: "bar" },
    width: "container",
    height: 400,
    autosize: { type: "fit", resize: true },
    encoding: {
      x: {
        field: "time",
        timeUnit: "utcyearmonthdatehours",
        type: "temporal",
        scale: {
          type: "utc",
          domain: [getHourlyFrom().getTime(), getHourlyTo().getTime()],
        },
        axis: {
          format: "%b %d",
          tickCount: 7,
          title: "",
        },
      },
      y: {
        field: "writeRecords",
        type: "quantitative",
        axis: {
          title: "",
        },
      },
      tooltip: [
        { field: "time", type: "temporal", title: "Hour Start", format: "%b %d %H:%M" },
        { field: "readRecords", type: "quantitative", title: "Read rows", format: "," },
        { field: "readBytes", type: "quantitative", title: "Read bytes", format: ".2s" },
        { field: "readOps", type: "quantitative", title: "Read ops", format: "," },
        { field: "writeRecords", type: "quantitative", title: "Write rows", format: "," },
        { field: "writeBytes", type: "quantitative", title: "Write bytes", format: ".2s" },
        { field: "writeOps", type: "quantitative", title: "Write ops", format: "," },
      ],
    },
    config: {
      arc: { fill: markColor },
      area: { fill: markColor },
      axis: {
        tickColor: "transparent",
        domain: false,
        gridDash: [5],
        gridColor: "rgba(255, 255, 255, 0.25)",
        labelFont: "-apple-system,BlinkMacSystemFont,Segoe UI,Roboto,Helvetica Neue,Ubuntu,sans-serif",
        labelFontSize: 14,
        labelColor: "rgba(255, 255, 255, 0.9)",
      },
      axisBand: { grid: false },
      background: "rgba(27, 35, 65, 1)",
      group: { fill: "rgba(35, 47, 74, 1)" },
      line: { stroke: markColor, strokeWidth: 2 },
      path: { stroke: markColor, strokeWidth: 0.5 },
      rect: { fill: markColor },
      range: {
        category: [
          "rgb(60, 170, 255)", //  blue
          "rgb(255, 60, 130)", //  red
          "rgb(100, 200, 100)", // green
          "rgb(200, 220, 240)", // gray
          "rgb(251, 241, 67)", //  yellow
          "rgb(203, 60, 255)", // purple
          "rgb(249, 148, 66)", // orange
          "rgb(36, 130, 36)", // dark green
          "rgb(238, 255, 234)", // pale green
        ],
        diverging: ["#cc0020", "#e77866", "#f6e7e1", "#d6e8ed", "#91bfd9", "#1d78b5"],
        heatmap: ["#d6e8ed", "#cee0e5", "#91bfd9", "#549cc6", "#1d78b5"],
      },
      point: { filled: true, shape: "circle" },
      shape: { stroke: markColor },
      symbol: { fill: markColor },
      style: {
        bar: {
          fill: "rgb(60, 170, 255)",
          stroke: "transparent",
          binSpacing: 0,
        },
      },
      title: {
        align: "left",
        color: "rgba(255, 255, 255, 0.9)",
        fontSize: 20,
        anchor: "start",
        fontWeight: 400,
        subtitlePadding: 50,
      },
      view: {
        stroke: "transparent",
      },
    },
  },
});

interface MetricsWithTime extends Metrics {
  time: Date;
}

const MetricsChart: FC<{ metrics: MetricsWithTime[] }> = ({ metrics }) => {
  const classes = useStyles();
  return (
    <Grid item xs={12}>
      <Paper className={classes.numberContainer}>
        <Typography className={classes.chartTitle}>ROWS WRITTEN IN THE LAST 7 DAYS</Typography>
        <MetricsChartVega
          data={{ metrics }}
          className={classes.vegaChart}
          actions={false}
          renderer="svg"
          // tooltip={(handler, event, item, value) => {
          //   console.log("here: ", item);
          // }}
        />
      </Paper>
    </Grid>
  );
};

const StreamMetricsOverview: FC<QueryStream> = ({ stream }) => {
  const { loading, error, data } = useQuery<GetStreamMetrics, GetStreamMetricsVariables>(GET_STREAM_METRICS, {
    variables: {
      streamID: stream.streamID,
      period: "M",
      from: getMonthFrom(),
    },
  });

  let total = blankMetrics();
  let current = blankMetrics();
  if (data && data.getStreamMetrics && data.getStreamMetrics.length > 0) {
    total = computeAggregateMetrics(data.getStreamMetrics);
    current = data.getStreamMetrics[data.getStreamMetrics.length - 1];
  }

  return (
    <>
      <MetricsOverview current={current} total={total} alltime />
      {error && <ErrorNote error={error} />}
    </>
  );
};

const StreamMetricsWeek: FC<QueryStream> = ({ stream }) => {
  const { loading, error, data } = useQuery<GetStreamMetrics, GetStreamMetricsVariables>(GET_STREAM_METRICS, {
    variables: {
      streamID: stream.streamID,
      period: "H",
      from: getHourlyFrom(),
    },
  });

  let total = blankMetrics();
  let current = blankMetrics();
  if (data && data.getStreamMetrics && data.getStreamMetrics.length > 0) {
    total = computeAggregateMetrics(data.getStreamMetrics);
    current = data.getStreamMetrics[data.getStreamMetrics.length - 1];
  }

  let metrics: MetricsWithTime[];
  if (data && data.getStreamMetrics.length > 0) {
    metrics = data.getStreamMetrics;
  } else {
    metrics = [{ time: getHourlyFrom(), ...current }, { time: getHourlyTo(), ...current }];
  }

  return (
    <>
      <MetricsChart metrics={metrics} />
      <MetricsOverview current={current} total={total} alltime={false} />
      {error && <ErrorNote error={error} />}
    </>
  );
};

const StreamMetrics: FC<QueryStream> = ({ stream }) => {
  return (
    <Grid container spacing={2}>
      <StreamMetricsOverview stream={stream} />
      <StreamMetricsWeek stream={stream} />
    </Grid>
  );
};

export default StreamMetrics;
