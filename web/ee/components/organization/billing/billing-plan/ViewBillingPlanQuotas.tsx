import React, { FC } from "react";

import { BillingInfo_billingInfo_billingPlan } from "ee/apollo/types/BillingInfo";
import { Grid, makeStyles, Table, TableBody, TableCell, TableHead, TableRow, Typography } from "@material-ui/core";
import { prettyPrintBytes } from "components/metrics/util";
import VSpace from "components/VSpace";

const useStyles = makeStyles((theme) => ({
  paperPadding: {
    padding: theme.spacing(3)
  },
  container: {
    overflowX: "auto"
  },
  tableKeyColumn: {
    backgroundColor: theme.palette.background?.medium,
    fontWeight: 500
  }
}));

interface Props {
  billingPlan: BillingInfo_billingInfo_billingPlan;
}

const ViewBillingPlanQuotas: FC<Props> = ({billingPlan}) => {
  const classes = useStyles();
  const currencyFormatter = new Intl.NumberFormat('en-US', {style: 'currency', currency: 'USD'});

  const isOverage = (billingPlan.readOveragePriceCents > 0) || (billingPlan.writeOveragePriceCents > 0) || (billingPlan.scanOveragePriceCents > 0);

  let allowedReadOverage;
  let allowedWriteOverage;
  let allowedScanOverage;
  let maximumReadOverageCharge;
  let maximumWriteOverageCharge;
  let maximumScanOverageCharge;
  if (isOverage) {
    allowedReadOverage = billingPlan.readQuota - billingPlan.baseReadQuota;
    allowedWriteOverage = billingPlan.writeQuota - billingPlan.baseWriteQuota;
    allowedScanOverage = billingPlan.scanQuota - billingPlan.baseScanQuota;
    maximumReadOverageCharge = allowedReadOverage / 10**9 * billingPlan.readOveragePriceCents;
    maximumWriteOverageCharge = allowedWriteOverage / 10**9 * billingPlan.writeOveragePriceCents;
    maximumScanOverageCharge = allowedScanOverage / 10**9 * billingPlan.scanOveragePriceCents;
  }

  return (
    <>
      <Grid container className={classes.container}>
        <Grid item>
            <Typography variant="h3">Quotas</Typography>
            <VSpace units={2} />
            {!isOverage && (
              <>
              <Grid container>
                <Grid item>
                  <Table size="medium">
                    <TableHead>
                      <TableRow>
                        <TableCell></TableCell>
                        <TableCell align="right">Quota</TableCell>
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      <TableRow>
                        <TableCell className={classes.tableKeyColumn}>Reads</TableCell>
                        <TableCell align="right">{prettyPrintBytes(billingPlan.baseReadQuota)}</TableCell>
                      </TableRow>
                      <TableRow>
                        <TableCell className={classes.tableKeyColumn}>Writes</TableCell>
                        <TableCell align="right">{prettyPrintBytes(billingPlan.baseWriteQuota)}</TableCell>
                      </TableRow>
                      <TableRow>
                        <TableCell className={classes.tableKeyColumn}>Scans</TableCell>
                        <TableCell align="right">{prettyPrintBytes(billingPlan.baseScanQuota)}</TableCell>
                      </TableRow>
                    </TableBody>
                  </Table>
                </Grid>
              </Grid>
              </>
            )}
            {isOverage && (
              <>
                <Grid container>
                  <Grid item>
                    <Table size="medium">
                      <TableHead>
                        <TableRow>
                          <TableCell></TableCell>
                          <TableCell align="right">Prepaid quota</TableCell>
                          <TableCell align="right">Allowed overage</TableCell>
                          <TableCell align="right">Price per overage GB</TableCell>
                          <TableCell align="right">Maximum overage charge</TableCell>
                        </TableRow>
                      </TableHead>
                      <TableBody>
                        <TableRow>
                          <TableCell className={classes.tableKeyColumn}>Reads</TableCell>
                          <TableCell align="right">{prettyPrintBytes(billingPlan.baseReadQuota)}</TableCell>
                          <TableCell align="right">{prettyPrintBytes(allowedReadOverage as number)}</TableCell>
                          <TableCell align="right">{currencyFormatter.format(billingPlan.readOveragePriceCents / 100)}</TableCell>
                          <TableCell align="right">{currencyFormatter.format(maximumReadOverageCharge as number / 100)}</TableCell>
                        </TableRow>
                        <TableRow>
                          <TableCell className={classes.tableKeyColumn}>Writes</TableCell>
                          <TableCell align="right">{prettyPrintBytes(billingPlan.baseWriteQuota)}</TableCell>
                          <TableCell align="right">{prettyPrintBytes(allowedWriteOverage as number)}</TableCell>
                          <TableCell align="right">{currencyFormatter.format(billingPlan.writeOveragePriceCents / 100)}</TableCell>
                          <TableCell align="right">{currencyFormatter.format(maximumWriteOverageCharge as number / 100)}</TableCell>
                        </TableRow>
                        <TableRow>
                          <TableCell className={classes.tableKeyColumn}>Scans</TableCell>
                          <TableCell align="right">{prettyPrintBytes(billingPlan.baseScanQuota)}</TableCell>
                          <TableCell align="right">{prettyPrintBytes(allowedScanOverage as number)}</TableCell>
                          <TableCell align="right">{currencyFormatter.format(billingPlan.scanOveragePriceCents / 100)}</TableCell>
                          <TableCell align="right">{currencyFormatter.format(maximumScanOverageCharge as number / 100)}</TableCell>
                        </TableRow>
                      </TableBody>
                    </Table>
                  </Grid>
                </Grid>
              </>
            )}
        </Grid>
      </Grid>
    </>
  );
};

export default ViewBillingPlanQuotas;