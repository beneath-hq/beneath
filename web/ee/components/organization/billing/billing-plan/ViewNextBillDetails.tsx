import { Grid, makeStyles, Table, TableBody, TableCell, TableHead, TableRow } from "@material-ui/core";
import { OrganizationByName_organizationByName_PrivateOrganization } from "apollo/types/OrganizationByName";
import { prettyPrintBytes } from "components/metrics/util";
import { BillingInfo_billingInfo } from "ee/apollo/types/BillingInfo";
import { FC } from "react";

const useStyles = makeStyles((theme) => ({
  container: {
    overflowX: "auto",
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
    return <></>;
  }

  const readOverageTotal = billingInfo.billingPlan.readOveragePriceCents * (organization.readUsage - organization.prepaidReadQuota > 0 ? organization.readUsage - organization.prepaidReadQuota : 0);
  const writeOverageTotal = billingInfo.billingPlan.writeOveragePriceCents * (organization.writeUsage - organization.prepaidWriteQuota > 0 ? organization.writeUsage - organization.prepaidWriteQuota : 0);
  const scanOverageTotal = billingInfo.billingPlan.scanOveragePriceCents * (organization.scanUsage - organization.prepaidScanQuota > 0 ? organization.scanUsage - organization.prepaidScanQuota : 0);

  return (
    <>
    <Grid container className={classes.container}>
      <Grid item xs={12}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Operation</TableCell>
              <TableCell align="right">Prepaid Quota</TableCell>
              <TableCell align="right">Allowed Overage</TableCell>
              <TableCell align="right">Used</TableCell>
              <TableCell align="right">Price per overage GB</TableCell>
              <TableCell align="right">Overage charge</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            <TableRow>
              <TableCell>Reads</TableCell>
              <TableCell align="right">{prettyPrintBytes(organization.prepaidReadQuota)}</TableCell>
              <TableCell align="right">{prettyPrintBytes(organization.readQuota - organization.prepaidReadQuota)}</TableCell>
              <TableCell align="right">{prettyPrintBytes(organization.readUsage)}</TableCell>
              <TableCell align="right">{currencyFormatter.format(billingInfo.billingPlan.readOveragePriceCents / 100)}</TableCell>
              <TableCell align="right">{currencyFormatter.format(readOverageTotal / 100)}</TableCell>
            </TableRow>
            <TableRow>
              <TableCell>Writes</TableCell>
              <TableCell align="right">{prettyPrintBytes(organization.prepaidWriteQuota)}</TableCell>
              <TableCell align="right">{prettyPrintBytes(organization.writeQuota - organization.prepaidWriteQuota)}</TableCell>
              <TableCell align="right">{prettyPrintBytes(organization.writeUsage)}</TableCell>
              <TableCell align="right">{currencyFormatter.format(billingInfo.billingPlan.writeOveragePriceCents / 100)}</TableCell>
              <TableCell align="right">{currencyFormatter.format(writeOverageTotal / 100)}</TableCell>
            </TableRow>
            <TableRow>
              <TableCell>Scans</TableCell>
              <TableCell align="right">{prettyPrintBytes(organization.prepaidScanQuota)}</TableCell>
              <TableCell align="right">{prettyPrintBytes(organization.scanQuota - organization.prepaidScanQuota)}</TableCell>
              <TableCell align="right">{prettyPrintBytes(organization.scanUsage)}</TableCell>
              <TableCell align="right">{currencyFormatter.format(billingInfo.billingPlan.scanOveragePriceCents / 100)}</TableCell>
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