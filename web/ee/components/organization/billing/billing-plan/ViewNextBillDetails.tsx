import { Grid, makeStyles, Table, TableBody, TableCell, TableHead, TableRow, Typography } from "@material-ui/core";
import { FC } from "react";

import { OrganizationByName_organizationByName_PrivateOrganization } from "apollo/types/OrganizationByName";
import { prettyPrintBytes } from "components/metrics/util";
import { BillingInfo_billingInfo } from "ee/apollo/types/BillingInfo";

const useStyles = makeStyles((theme) => ({
  sectionHeader: {
    marginTop: theme.spacing(3),
    marginBottom: theme.spacing(3)
  },
  container: {
    overflowX: "auto",
  },
  tableKeyColumn: {
    backgroundColor: theme.palette.background?.medium,
    fontWeight: 500
  }
}));

interface Props {
  organization: OrganizationByName_organizationByName_PrivateOrganization;
  billingInfo: BillingInfo_billingInfo;
}

const ViewNextBillDetails: FC<Props> = ({organization, billingInfo}) => {
  const classes = useStyles();
  const currencyFormatter = new Intl.NumberFormat('en-US', {style: 'currency', currency: 'USD'});

  if (!organization.readQuota || !organization.prepaidReadQuota || !organization.writeQuota || !organization.prepaidWriteQuota || !organization.scanQuota || !organization.prepaidScanQuota) {
    return null;
  }

  const readOverageTotal = billingInfo.billingPlan.readOveragePriceCents * (organization.readUsage - organization.prepaidReadQuota > 0 ? organization.readUsage - organization.prepaidReadQuota : 0);
  const writeOverageTotal = billingInfo.billingPlan.writeOveragePriceCents * (organization.writeUsage - organization.prepaidWriteQuota > 0 ? organization.writeUsage - organization.prepaidWriteQuota : 0);
  const scanOverageTotal = billingInfo.billingPlan.scanOveragePriceCents * (organization.scanUsage - organization.prepaidScanQuota > 0 ? organization.scanUsage - organization.prepaidScanQuota : 0);

  return (
    <>
      <Typography variant="h2" className={classes.sectionHeader} >
        Details of your next bill
      </Typography>

      <Grid container className={classes.container}>
        <Grid item xs={12}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell></TableCell>
                <TableCell align="right">Prepaid quota</TableCell>
                <TableCell align="right">Allowed overage</TableCell>
                <TableCell align="right">Price per overage GB</TableCell>
                <TableCell align="right">Your usage</TableCell>
                <TableCell align="right">Your overage cost</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              <TableRow>
                <TableCell className={classes.tableKeyColumn}>Reads</TableCell>
                <TableCell align="right">{prettyPrintBytes(organization.prepaidReadQuota)}</TableCell>
                <TableCell align="right">{prettyPrintBytes(organization.readQuota - organization.prepaidReadQuota)}</TableCell>
                <TableCell align="right">{currencyFormatter.format(billingInfo.billingPlan.readOveragePriceCents / 100)}</TableCell>
                <TableCell align="right">{prettyPrintBytes(organization.readUsage)}</TableCell>
                <TableCell align="right">{currencyFormatter.format(readOverageTotal / 100)}</TableCell>
              </TableRow>
              <TableRow>
                <TableCell className={classes.tableKeyColumn}>Writes</TableCell>
                <TableCell align="right">{prettyPrintBytes(organization.prepaidWriteQuota)}</TableCell>
                <TableCell align="right">{prettyPrintBytes(organization.writeQuota - organization.prepaidWriteQuota)}</TableCell>
                <TableCell align="right">{currencyFormatter.format(billingInfo.billingPlan.writeOveragePriceCents / 100)}</TableCell>
                <TableCell align="right">{prettyPrintBytes(organization.writeUsage)}</TableCell>
                <TableCell align="right">{currencyFormatter.format(writeOverageTotal / 100)}</TableCell>
              </TableRow>
              <TableRow>
                <TableCell className={classes.tableKeyColumn}>Scans</TableCell>
                <TableCell align="right">{prettyPrintBytes(organization.prepaidScanQuota)}</TableCell>
                <TableCell align="right">{prettyPrintBytes(organization.scanQuota - organization.prepaidScanQuota)}</TableCell>
                <TableCell align="right">{currencyFormatter.format(billingInfo.billingPlan.scanOveragePriceCents / 100)}</TableCell>
                <TableCell align="right">{prettyPrintBytes(organization.scanUsage)}</TableCell>
                <TableCell align="right">{currencyFormatter.format(scanOverageTotal / 100)}</TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </Grid>
      </Grid>
    </>
  );
};

export default ViewNextBillDetails;