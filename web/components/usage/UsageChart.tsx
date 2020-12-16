import { FC } from "react";
import { makeStyles, Theme } from "@material-ui/core";
import { VegaLite } from "react-vega";

import { Usage } from "./hooks";
import { vegaConfig } from "lib/vega";

const useStyles = makeStyles((theme: Theme) => ({
  chart: {
    width: "100%",
  },
}));

export type UsageUnit = "bytes" | "ops" | "records";
export type UsageDimension = "read" | "write" | "scan";

export interface UsageChartProps {
  usages: Usage[];
  unit: UsageUnit;
  dimension: UsageDimension;
}

export const UsageChart: FC<UsageChartProps> = ({ usages, unit, dimension }) => {
  let y = "";
  if (dimension === "read") {
    if (unit === "bytes") {
      y = "readBytes";
    } else if (unit === "ops") {
      y = "readOps";
    } else if (unit === "records") {
      y = "readRecords";
    }
  } else if (dimension === "write") {
    if (unit === "bytes") {
      y = "writeBytes";
    } else if (unit === "ops") {
      y = "writeOps";
    } else if (unit === "records") {
      y = "writeRecords";
    }
  } else if (dimension === "scan") {
    if (unit === "bytes") {
      y = "scanBytes";
    } else if (unit === "ops") {
      y = "scanOps";
    }
  }
  
  const yIsBytes = y === "readBytes" || y === "writeBytes" || y === "scanBytes";
  const classes = useStyles();
  return (
    <VegaLite
      className={classes.chart}
      data={{ usages }}
      actions={false}
      renderer="svg"
      // onNewView={(view) => {}} // HINT: To do SSR, must not return until this has triggered
      spec={{
        $schema: "https://vega.github.io/schema/vega-lite/v4.0.0-beta.10.json",
        config: vegaConfig,
        height: 400,
        width: "container",
        autosize: { type: "fit", resize: true },
        data: { name: "usages" },
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
  );
};

export default UsageChart;
