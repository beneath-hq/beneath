import { FC } from "react";
import { makeStyles, Theme } from "@material-ui/core";
import { VegaLite } from "react-vega";

import { Usage, UsageUnit, UsageDimension, usageFieldFor } from "./util";
import { vegaConfig } from "lib/vega";

const useStyles = makeStyles((theme: Theme) => ({
  chart: {
    width: "100%",
  },
}));

export interface UsageChartProps {
  usages: Usage[];
  unit: UsageUnit;
  dimension: UsageDimension;
}

export const UsageChart: FC<UsageChartProps> = ({ usages, unit, dimension }) => {
  const y = usageFieldFor(unit, dimension);
  const yIsBytes = unit === "bytes";
  const minY = yIsBytes ? 1000 : 10;
  const classes = useStyles();
  return (
    <VegaLite
      className={classes.chart}
      actions={false}
      renderer="svg"
      // onNewView={(view) => {}} // HINT: To do SSR, must not return until this has triggered
      spec={{
        $schema: "https://vega.github.io/schema/vega-lite/v5.1.0.json",
        config: vegaConfig,
        height: 400,
        width: "container",
        autosize: { type: "fit", resize: true },
        data: { values: usages },
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
            { field: "scanBytes", type: "quantitative", title: "Scan bytes", format: ".2s" },
            { field: "scanOps", type: "quantitative", title: "Scan ops", format: "," },
          ],
        },
        layer: [
          {
            mark: {
              type: "line",
            },
            encoding: {
              y: {
                field: y,
                type: "quantitative",
                axis: {
                  title: "",
                  format: yIsBytes ? "~s" : undefined,
                  labelExpr: yIsBytes ? "datum.label + 'B'" : undefined,
                },
                scale: {
                  domain: { unionWith: [0, minY] },
                },
              },
            },
          },
          {
            mark: "rule",
            params: [
              {
                name: "hover",
                select: {
                  type: "point",
                  fields: ["time"],
                  nearest: true,
                  on: "mouseover",
                  clear: "mouseout",
                },
              },
            ],
            encoding: {
              color: {
                value: "transparent",
                condition: {
                  param: "hover",
                  value: "white",
                  empty: false,
                },
              },
            },
          },
        ],
      }}
    />
  );
};

export default UsageChart;
