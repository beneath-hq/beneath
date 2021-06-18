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
import numbro from "numbro";
import { FC } from "react";

import { OrganizationByName_organizationByName_PrivateOrganization } from "apollo/types/OrganizationByName";
import { BillingInfo_billingInfo } from "ee/apollo/types/BillingInfo";

const bytesFormat: numbro.Format = { base: "decimal", mantissa: 1, optionalMantissa: true, output: "byte" };

const useStyles = makeStyles((theme) => ({
  sectionHeader: {
    marginTop: theme.spacing(3),
    marginBottom: theme.spacing(3),
  },
  container: {
    overflowX: "auto",
  },
  keyColumn: {
    backgroundColor: theme.palette.background?.medium,
    fontWeight: 500,
  },
}));

interface Props {
  organization: OrganizationByName_organizationByName_PrivateOrganization;
  billingInfo: BillingInfo_billingInfo;
}

const ViewNextBillDetails: FC<Props> = ({ organization, billingInfo }) => {
  const classes = useStyles();
  const currencyFormatter = new Intl.NumberFormat("en-US", { style: "currency", currency: "USD" });

  if (
    !organization.readQuota ||
    !organization.prepaidReadQuota ||
    !organization.writeQuota ||
    !organization.prepaidWriteQuota ||
    !organization.scanQuota ||
    !organization.prepaidScanQuota
  ) {
    return null;
  }

  const readOverageTotal =
    billingInfo.billingPlan.readOveragePriceCents *
    (organization.readUsage - organization.prepaidReadQuota > 0
      ? organization.readUsage - organization.prepaidReadQuota
      : 0);
  const writeOverageTotal =
    billingInfo.billingPlan.writeOveragePriceCents *
    (organization.writeUsage - organization.prepaidWriteQuota > 0
      ? organization.writeUsage - organization.prepaidWriteQuota
      : 0);
  const scanOverageTotal =
    billingInfo.billingPlan.scanOveragePriceCents *
    (organization.scanUsage - organization.prepaidScanQuota > 0
      ? organization.scanUsage - organization.prepaidScanQuota
      : 0);

  return (
    <>
      <Typography variant="h2" className={classes.sectionHeader}>
        Details of your next bill
      </Typography>

      <Grid container className={classes.container}>
        <Grid item xs={12}>
          <MuiTable>
            <MuiTableHead>
              <MuiTableRow>
                <MuiTableCell></MuiTableCell>
                <MuiTableCell align="right">Prepaid quota</MuiTableCell>
                <MuiTableCell align="right">Allowed overage</MuiTableCell>
                <MuiTableCell align="right">Price per overage GB</MuiTableCell>
                <MuiTableCell align="right">Your usage</MuiTableCell>
                <MuiTableCell align="right">Your overage cost</MuiTableCell>
              </MuiTableRow>
            </MuiTableHead>
            <MuiTableBody>
              <MuiTableRow>
                <MuiTableCell className={classes.keyColumn}>Reads</MuiTableCell>
                <MuiTableCell align="right">{numbro(organization.prepaidReadQuota).format(bytesFormat)}</MuiTableCell>
                <MuiTableCell align="right">
                  {numbro(organization.readQuota - organization.prepaidReadQuota).format(bytesFormat)}
                </MuiTableCell>
                <MuiTableCell align="right">
                  {currencyFormatter.format(billingInfo.billingPlan.readOveragePriceCents / 100)}
                </MuiTableCell>
                <MuiTableCell align="right">{numbro(organization.readUsage).format(bytesFormat)}</MuiTableCell>
                <MuiTableCell align="right">{currencyFormatter.format(readOverageTotal / 100)}</MuiTableCell>
              </MuiTableRow>
              <MuiTableRow>
                <MuiTableCell className={classes.keyColumn}>Writes</MuiTableCell>
                <MuiTableCell align="right">{numbro(organization.prepaidWriteQuota).format(bytesFormat)}</MuiTableCell>
                <MuiTableCell align="right">
                  {numbro(organization.writeQuota - organization.prepaidWriteQuota).format(bytesFormat)}
                </MuiTableCell>
                <MuiTableCell align="right">
                  {currencyFormatter.format(billingInfo.billingPlan.writeOveragePriceCents / 100)}
                </MuiTableCell>
                <MuiTableCell align="right">{numbro(organization.writeUsage).format(bytesFormat)}</MuiTableCell>
                <MuiTableCell align="right">{currencyFormatter.format(writeOverageTotal / 100)}</MuiTableCell>
              </MuiTableRow>
              <MuiTableRow>
                <MuiTableCell className={classes.keyColumn}>Scans</MuiTableCell>
                <MuiTableCell align="right">{numbro(organization.prepaidScanQuota).format(bytesFormat)}</MuiTableCell>
                <MuiTableCell align="right">
                  {numbro(organization.scanQuota - organization.prepaidScanQuota).format(bytesFormat)}
                </MuiTableCell>
                <MuiTableCell align="right">
                  {currencyFormatter.format(billingInfo.billingPlan.scanOveragePriceCents / 100)}
                </MuiTableCell>
                <MuiTableCell align="right">{numbro(organization.scanUsage).format(bytesFormat)}</MuiTableCell>
                <MuiTableCell align="right">{currencyFormatter.format(scanOverageTotal / 100)}</MuiTableCell>
              </MuiTableRow>
            </MuiTableBody>
          </MuiTable>
        </Grid>
      </Grid>
    </>
  );
};

export default ViewNextBillDetails;
