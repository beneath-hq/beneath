import { Config } from "vega-lite";
import theme from "./theme";

const colors = {
  mark: "rgb(60, 170, 255)",
  categorical: [
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
};

export const vegaConfig: Config = {
  // arc: { fill: colors.mark },
  area: { fill: colors.mark },
  axis: {
    domain: false,
    grid: false,
    labelFont: theme.typography.fontFamily,
    labelFontSize: 14,
    labelColor: theme.palette.text.secondary,
    tickColor: "transparent",
  },
  axisY: {
    grid: true,
    gridDash: [5],
    gridColor: "rgba(255, 255, 255, 0.25)",
  },
  axisBand: { grid: false },
  background: theme.palette.background.paper,
  // group: { fill: theme.palette.background.paper },
  line: { stroke: colors.mark, strokeWidth: 2 },
  // path: { stroke: colors.mark, strokeWidth: 0.5 },
  rect: { fill: colors.mark },
  range: {
    category: colors.categorical,
    diverging: colors.diverging,
    heatmap: colors.heatmap,
  },
  point: { filled: false, shape: "circle", stroke: colors.mark, fill: theme.palette.background.paper },
  // shape: { stroke: colors.mark },
  // symbol: { fill: colors.mark },
  rule: { stroke: "rgba(255, 255, 255, 0.5)" },
  style: {
    bar: {
      fill: colors.mark,
      stroke: "transparent",
    },
  },
  title: {
    align: "left",
    anchor: "start",
    color: theme.palette.text.primary,
    fontSize: 20,
    fontWeight: "bold",
  },
  view: {
    stroke: "transparent",
  },
};
