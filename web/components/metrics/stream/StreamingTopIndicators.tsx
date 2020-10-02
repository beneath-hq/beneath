import numbro from "numbro";
import { FC } from "react";

import { Grid, Paper } from "@material-ui/core";

import SingleIndicator from "../SingleIndicator";
import { Metrics } from "../util";

const intFormat = { thousandSeparated: true };
const bytesFormat: numbro.Format = { base: "decimal", mantissa: 1, output: "byte" };

export interface TopIndicatorsProps {
  latest: Metrics;
  total: Metrics;
  period: "hour" | "month";
  totalPeriod: "week" | "all time";
}

export const StreamingTopIndicators: FC<TopIndicatorsProps> = ({ latest, total, period, totalPeriod }) => {
  const formatTitle = (msg: string) => (totalPeriod === "all time" ? `total ${msg}` : `${msg} this ${totalPeriod}`);
  const formatNote = (msg: string) => `(${msg} this ${period})`;
  return (
    <Grid container spacing={2} item xs={12}>
      <Grid item xs={6} md={3}>
        <SingleIndicator
          title={formatTitle("Records Written")}
          indicator={numbro(total.writeRecords).format(intFormat)}
          note={formatNote(numbro(latest.writeRecords).format(intFormat))}
        />
      </Grid>
      <Grid item xs={6} md={3}>
        <Paper>
          <SingleIndicator
            title={formatTitle("Bytes Written")}
            indicator={numbro(total.writeBytes).format(bytesFormat)}
            note={formatNote(numbro(latest.writeBytes).format(bytesFormat))}
          />
        </Paper>
      </Grid>
      <Grid item xs={6} md={3}>
        <Paper>
          <SingleIndicator
            title={formatTitle("Read Calls")}
            indicator={numbro(total.readOps).format(intFormat)}
            note={formatNote(numbro(latest.readOps).format(intFormat))}
          />
        </Paper>
      </Grid>
      <Grid item xs={6} md={3}>
        <Paper>
          <SingleIndicator
            title={formatTitle("Records Read")}
            indicator={numbro(total.readRecords).format(intFormat)}
            note={formatNote(numbro(latest.readRecords).format(intFormat))}
          />
        </Paper>
      </Grid>
    </Grid>
  );
};

export default StreamingTopIndicators;
