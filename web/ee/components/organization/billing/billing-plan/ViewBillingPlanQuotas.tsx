import numbro from "numbro";
import React, { FC } from "react";

import { BillingInfo_billingInfo_billingPlan } from "ee/apollo/types/BillingInfo";
import {
  Grid,
  makeStyles,
  Table as MuiTable,
  TableBody as MuiTableBody,
  TableCell as MuiTableCell,
  TableHead as MuiTableHead,
  TableRow as MuiTableRow,
  Typography,
} from "@material-ui/core";
import VSpace from "components/VSpace";

const bytesFormat: numbro.Format = { base: "decimal", mantissa: 0, output: "byte" };

const useStyles = makeStyles((theme) => ({
  paperPadding: {
    padding: theme.spacing(3),
  },
  container: {
    overflowX: "auto",
  },
  keyColumn: {
    backgroundColor: theme.palette.background?.medium,
    width: theme.spacing(15),
    fontWeight: 500,
  },
}));

interface Props {
  billingPlan: BillingInfo_billingInfo_billingPlan;
}

const ViewBillingPlanQuotas: FC<Props> = ({ billingPlan }) => {
  const classes = useStyles();
  const currencyFormatter = new Intl.NumberFormat("en-US", { style: "currency", currency: "USD" });

  const isOverage =
    billingPlan.readOveragePriceCents > 0 ||
    billingPlan.writeOveragePriceCents > 0 ||
    billingPlan.scanOveragePriceCents > 0;

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
    maximumReadOverageCharge = (allowedReadOverage / 10 ** 9) * billingPlan.readOveragePriceCents;
    maximumWriteOverageCharge = (allowedWriteOverage / 10 ** 9) * billingPlan.writeOveragePriceCents;
    maximumScanOverageCharge = (allowedScanOverage / 10 ** 9) * billingPlan.scanOveragePriceCents;
  }

  return (
    <>
      <Typography variant="h3">Quotas</Typography>
      <VSpace units={2} />
      <Grid container className={classes.container}>
        <Grid item xs={12}>
          {!isOverage && (
            <>
              <MuiTable>
                <MuiTableHead>
                  <MuiTableRow>
                    <MuiTableCell></MuiTableCell>
                    <MuiTableCell align="center">Quota</MuiTableCell>
                  </MuiTableRow>
                </MuiTableHead>
                <MuiTableBody>
                  <MuiTableRow>
                    <MuiTableCell align="center" className={classes.keyColumn}>
                      Reads
                    </MuiTableCell>
                    <MuiTableCell align="center">{numbro(billingPlan.baseReadQuota).format(bytesFormat)}</MuiTableCell>
                  </MuiTableRow>
                  <MuiTableRow>
                    <MuiTableCell align="center" className={classes.keyColumn}>
                      Writes
                    </MuiTableCell>
                    <MuiTableCell align="center">{numbro(billingPlan.baseWriteQuota).format(bytesFormat)}</MuiTableCell>
                  </MuiTableRow>
                  <MuiTableRow>
                    <MuiTableCell align="center" className={classes.keyColumn}>
                      Scans
                    </MuiTableCell>
                    <MuiTableCell align="center">{numbro(billingPlan.baseScanQuota).format(bytesFormat)}</MuiTableCell>
                  </MuiTableRow>
                </MuiTableBody>
              </MuiTable>
            </>
          )}
          {isOverage && (
            <>
              <MuiTable>
                <MuiTableHead>
                  <MuiTableRow>
                    <MuiTableCell></MuiTableCell>
                    <MuiTableCell align="right">Prepaid quota</MuiTableCell>
                    <MuiTableCell align="right">Allowed overage</MuiTableCell>
                    <MuiTableCell align="right">Price per overage GB</MuiTableCell>
                    <MuiTableCell align="right">Maximum overage charge</MuiTableCell>
                  </MuiTableRow>
                </MuiTableHead>
                <MuiTableBody>
                  <MuiTableRow>
                    <MuiTableCell align="center" className={classes.keyColumn}>
                      Reads
                    </MuiTableCell>
                    <MuiTableCell align="right">{numbro(billingPlan.baseReadQuota).format(bytesFormat)}</MuiTableCell>
                    <MuiTableCell align="right">
                      {numbro(allowedReadOverage as number).format(bytesFormat)}
                    </MuiTableCell>
                    <MuiTableCell align="right">
                      {currencyFormatter.format(billingPlan.readOveragePriceCents / 100)}
                    </MuiTableCell>
                    <MuiTableCell align="right">
                      {currencyFormatter.format((maximumReadOverageCharge as number) / 100)}
                    </MuiTableCell>
                  </MuiTableRow>
                  <MuiTableRow>
                    <MuiTableCell align="center" className={classes.keyColumn}>
                      Writes
                    </MuiTableCell>
                    <MuiTableCell align="right">{numbro(billingPlan.baseWriteQuota).format(bytesFormat)}</MuiTableCell>
                    <MuiTableCell align="right">
                      {numbro(allowedWriteOverage as number).format(bytesFormat)}
                    </MuiTableCell>
                    <MuiTableCell align="right">
                      {currencyFormatter.format(billingPlan.writeOveragePriceCents / 100)}
                    </MuiTableCell>
                    <MuiTableCell align="right">
                      {currencyFormatter.format((maximumWriteOverageCharge as number) / 100)}
                    </MuiTableCell>
                  </MuiTableRow>
                  <MuiTableRow>
                    <MuiTableCell align="center" className={classes.keyColumn}>
                      Scans
                    </MuiTableCell>
                    <MuiTableCell align="right">{numbro(billingPlan.baseScanQuota).format(bytesFormat)}</MuiTableCell>
                    <MuiTableCell align="right">
                      {numbro(allowedScanOverage as number).format(bytesFormat)}
                    </MuiTableCell>
                    <MuiTableCell align="right">
                      {currencyFormatter.format(billingPlan.scanOveragePriceCents / 100)}
                    </MuiTableCell>
                    <MuiTableCell align="right">
                      {currencyFormatter.format((maximumScanOverageCharge as number) / 100)}
                    </MuiTableCell>
                  </MuiTableRow>
                </MuiTableBody>
              </MuiTable>
            </>
          )}
        </Grid>
      </Grid>
    </>
  );
};

export default ViewBillingPlanQuotas;
