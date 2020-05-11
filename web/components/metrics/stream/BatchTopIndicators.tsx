/*
- If batch
  - Show written records and bytes written of current batch
  - Show calls and records read of all batches
  - Show number of past instances committed
    - instancesCreatedCount
    - instancesCommittedCount
  - Show chart of reads
*/

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
  period: "month";
  instancesCreated: number;
  instancesCommitted: number;
}

export const StreamingTopIndicators: FC<TopIndicatorsProps> = ({
  latest,
  total,
  instancesCreated,
  instancesCommitted,
}) => {
  return (
    <Grid container spacing={2} item xs={12}>
      <Grid item xs={6} md={4}>
        <SingleIndicator title={"Rows in latest batch"} indicator={numbro(latest.writeRecords).format(intFormat)} />
      </Grid>
      <Grid item xs={6} md={4}>
        <Paper>
          <SingleIndicator title={"Data in latest batch"} indicator={numbro(total.writeBytes).format(bytesFormat)} />
        </Paper>
      </Grid>
      <Grid item xs={6} md={4}>
        <Paper>
          <SingleIndicator title={"Batches finalized"} indicator={numbro(instancesCommitted).format(intFormat)} />
        </Paper>
      </Grid>
      <Grid item xs={6} md={4}>
        <Paper>
          <SingleIndicator
            title={"Read calls (all batches)"}
            indicator={numbro(total.readOps).format(intFormat)}
            note={`(${numbro(latest.readOps).format(intFormat)} last month)`}
          />
        </Paper>
      </Grid>
      <Grid item xs={6} md={4}>
        <Paper>
          <SingleIndicator
            title={"Rows read (all batches)"}
            indicator={numbro(total.readRecords).format(intFormat)}
            note={`(${numbro(latest.readRecords).format(intFormat)} last month)`}
          />
        </Paper>
      </Grid>
      <Grid item xs={6} md={4}>
        <Paper>
          <SingleIndicator
            title={"Bytes read (all batches)"}
            indicator={numbro(total.readBytes).format(bytesFormat)}
            note={`(${numbro(latest.readBytes).format(bytesFormat)} last month)`}
          />
        </Paper>
      </Grid>
    </Grid>
  );
};

export default StreamingTopIndicators;
