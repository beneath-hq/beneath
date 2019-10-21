import numbro from "numbro";
import { FC } from "react";

import { Grid, Paper } from "@material-ui/core";

import SingleIndicator from "../SingleIndicator";
import { Metrics } from "../util";

const intFormat = { thousandSeparated: true };
const bytesFormat = { base: "decimal", mantissa: 1, output: "byte" };

export interface TopIndicatorsProps {
  latest: Metrics;
  total: Metrics;
}

export const TopIndicators: FC<TopIndicatorsProps> = ({ latest, total }) => {
  const formatTitle = (msg: string) => `${msg} this month`;
  const formatNote = (msg: string) => `(${msg} all-time)`;
  return (
    <Grid container spacing={2} item xs={12}>
      <Grid item xs={6} md={4}>
        <SingleIndicator
          title={formatTitle("Read calls")}
          indicator={numbro(latest.readOps).format(intFormat)}
          note={formatNote(numbro(total.readOps).format(intFormat))}
        />
      </Grid>
      <Grid item xs={6} md={4}>
        <Paper>
          <SingleIndicator
            title={formatTitle("Read rows")}
            indicator={numbro(latest.readRecords).format(intFormat)}
            note={formatNote(numbro(total.readRecords).format(intFormat))}
          />
        </Paper>
      </Grid>
      <Grid item xs={6} md={4}>
        <Paper>
          <SingleIndicator
            title={formatTitle("Read bytes")}
            indicator={numbro(latest.readBytes).format(bytesFormat)}
            note={formatNote(numbro(total.readBytes).format(bytesFormat))}
          />
        </Paper>
      </Grid>
      <Grid item xs={6} md={4}>
        <SingleIndicator
          title={formatTitle("Write calls")}
          indicator={numbro(latest.writeOps).format(intFormat)}
          note={formatNote(numbro(total.writeOps).format(intFormat))}
        />
      </Grid>
      <Grid item xs={6} md={4}>
        <Paper>
          <SingleIndicator
            title={formatTitle("Write rows")}
            indicator={numbro(latest.writeRecords).format(intFormat)}
            note={formatNote(numbro(total.writeRecords).format(intFormat))}
          />
        </Paper>
      </Grid>
      <Grid item xs={6} md={4}>
        <Paper>
          <SingleIndicator
            title={formatTitle("Write bytes")}
            indicator={numbro(latest.writeBytes).format(bytesFormat)}
            note={formatNote(numbro(total.writeBytes).format(bytesFormat))}
          />
        </Paper>
      </Grid>
    </Grid>
  );
};

export default TopIndicators;
