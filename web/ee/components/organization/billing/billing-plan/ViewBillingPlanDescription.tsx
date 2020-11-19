import React, { FC } from "react";

import { BillingInfo_billingInfo_billingPlan } from "ee/apollo/types/BillingInfo";
import { Grid, makeStyles, Paper, Table, TableBody, TableCell, TableHead, TableRow, Typography } from "@material-ui/core";
import { prettyPrintBytes } from "components/metrics/util";
import VSpace from "components/VSpace";
import { PROFESSIONAL_BOOST_PLAN, PROFESSIONAL_PLAN } from "ee/lib/billing";

const useStyles = makeStyles((theme) => ({
  paperPadding: {
    padding: theme.spacing(3)
  },
  container: {
    overflowX: "auto"
  }
}));

interface Props {
  billingPlan: BillingInfo_billingInfo_billingPlan;
}

const ViewBillingPlanDescription: FC<Props> = ({billingPlan}) => {
  const classes = useStyles();
  const currencyFormatter = new Intl.NumberFormat('en-US', {style: 'currency', currency: 'USD'});

  const isProfessional = (billingPlan.description === PROFESSIONAL_PLAN);
  const isProfessionalBoost = (billingPlan.description === PROFESSIONAL_BOOST_PLAN);

  let description;
  if (isProfessional) {
    description = "For professional data workers building meaningful data pipelines.";
  }

  let allowedReadOverage;
  let allowedWriteOverage;
  let allowedScanOverage;
  let maximumReadOverageCharge;
  let maximumWriteOverageCharge;
  let maximumScanOverageCharge;
  if (isProfessionalBoost) {
    description = "The Boost allows you to outgrow the Professional plan. You get extra capacity (aka 'overage') on a pay-per-GB basis.";
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
          <Paper className={classes.paperPadding} variant="outlined">
            <Grid container alignItems="center" spacing={2} justify="space-between">
              <Grid item>
                <Typography variant="h2">{billingPlan.description}</Typography>
              </Grid>
              <Grid item>
                <Typography>
                  {(billingPlan.description === PROFESSIONAL_BOOST_PLAN ? "starting at " : "") + currencyFormatter.format(billingPlan.basePriceCents / 100)} / month
                </Typography>
              </Grid>
            </Grid>
            <VSpace units={3} />
            <Typography>{description}</Typography>
            <VSpace units={5} />
            <Typography variant="h3">Quotas</Typography>
            <VSpace units={2} />
            {isProfessional && (
              <>
              <Grid container>
                <Grid item>
                  <Table size="small">
                    <TableHead>
                      <TableRow>
                        <TableCell>Operation</TableCell>
                        <TableCell align="right">Quota</TableCell>
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      <TableRow>
                        <TableCell>Reads</TableCell>
                        <TableCell align="right">{prettyPrintBytes(billingPlan.baseReadQuota)}</TableCell>
                      </TableRow>
                      <TableRow>
                        <TableCell>Writes</TableCell>
                        <TableCell align="right">{prettyPrintBytes(billingPlan.baseWriteQuota)}</TableCell>
                      </TableRow>
                      <TableRow>
                        <TableCell>Scans</TableCell>
                        <TableCell align="right">{prettyPrintBytes(billingPlan.baseScanQuota)}</TableCell>
                      </TableRow>
                    </TableBody>
                  </Table>
                </Grid>
              </Grid>
              </>
            )}
            {isProfessionalBoost && (
              <>
                <Grid container>
                  <Grid item>
                    <Table size="small">
                      <TableHead>
                        <TableRow>
                          <TableCell>Operation</TableCell>
                          <TableCell align="right">Prepaid Quota</TableCell>
                          <TableCell align="right">Allowed Overage</TableCell>
                          <TableCell align="right">Price per overage GB</TableCell>
                          <TableCell align="right">Maximum overage charge</TableCell>
                        </TableRow>
                      </TableHead>
                      <TableBody>
                        <TableRow>
                          <TableCell>Reads</TableCell>
                          <TableCell align="right">{prettyPrintBytes(billingPlan.baseReadQuota)}</TableCell>
                          <TableCell align="right">{prettyPrintBytes(allowedReadOverage as number)}</TableCell>
                          <TableCell align="right">{currencyFormatter.format(billingPlan.readOveragePriceCents / 100)}</TableCell>
                          <TableCell align="right">{currencyFormatter.format(maximumReadOverageCharge as number / 100)}</TableCell>
                        </TableRow>
                        <TableRow>
                          <TableCell>Writes</TableCell>
                          <TableCell align="right">{prettyPrintBytes(billingPlan.baseWriteQuota)}</TableCell>
                          <TableCell align="right">{prettyPrintBytes(allowedWriteOverage as number)}</TableCell>
                          <TableCell align="right">{currencyFormatter.format(billingPlan.writeOveragePriceCents / 100)}</TableCell>
                          <TableCell align="right">{currencyFormatter.format(maximumWriteOverageCharge as number / 100)}</TableCell>
                        </TableRow>
                        <TableRow>
                          <TableCell>Scans</TableCell>
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
          </Paper>
        </Grid>
      </Grid>
    </>
  );
};

export default ViewBillingPlanDescription;