import { FC } from "react";
import { VegaLite } from "react-vega";

import { Grid, makeStyles, Paper, Theme, Typography } from "@material-ui/core";

import { vegaConfig } from "../../lib/vega";
import { hourFloor, MetricsWithTime, now, weekAgo } from "./util";

export interface WeekChartProps {
  metrics: MetricsWithTime[];
  y1: string;
  title: string;
}

const useStyles = makeStyles((theme: Theme) => ({
  paper: {
    padding: theme.spacing(2),
  },
  chart: {
    width: "100%",
  },
  title: {
    fontSize: "1rem",
    fontWeight: "bold",
    padding: "5px 5px 10px 0px",
    textTransform: "uppercase",
  },
}));

const WeekChart: FC<WeekChartProps> = ({ metrics, y1, title }) => {
  const yIsBytes = (y1 === "readBytes" || y1 === "writeBytes");
  const classes = useStyles();
  return (
    <Grid container spacing={2} item xs={12}>
      <Grid item xs={12}>
        <Paper className={classes.paper}>
          <Typography className={classes.title}>{title}</Typography>
          <VegaLite
            className={classes.chart}
            data={{ metrics }}
            actions={false}
            renderer="svg"
            // onNewView={(view) => {}} // HINT: To do SSR, must not return until this has triggered
            spec={{
              $schema: "https://vega.github.io/schema/vega-lite/v4.0.0-beta.10.json",
              config: vegaConfig,
              height: 400,
              width: "container",
              autosize: { type: "fit", resize: true },
              data: { name: "metrics" },
              encoding: {
                x: {
                  field: "time",
                  timeUnit: "utcyearmonthdatehours",
                  type: "temporal",
                  scale: { type: "utc" },
                  axis: {
                    format: "%b %d",
                    tickCount: 7,
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
              layer: [
                {
                  mark: {
                    type: "line",
                    point: true,
                  },
                  encoding: {
                    y: {
                      field: y1,
                      type: "quantitative",
                      axis: {
                        title: "",
                        format: yIsBytes ? "~s" : undefined,
                        labelExpr: yIsBytes ? "datum.label + 'B'" : undefined,
                      },
                      scale: {
                        zero: true,
                      },
                    },
                  },
                },
                {
                  mark: "rule",
                  selection: {
                    hover: { type: "single", on: "mouseover", clear: "mouseout", empty: "none", nearest: true },
                  },
                  encoding: {
                    color: {
                      condition: {
                        selection: { not: "hover" },
                        value: "transparent",
                      },
                    },
                  },
                },
              ],
            }}
          />
        </Paper>
      </Grid>
    </Grid>
  );
};

export default WeekChart;
