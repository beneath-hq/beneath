import { useQuery } from "@apollo/client";
import _ from "lodash";
import { Button, Grid, Paper, Table, TableBody, TableCell, TableHead, TableRow, Typography } from "@material-ui/core";
import { makeStyles } from "@material-ui/core/styles";
import React, { FC } from "react";
import Moment from "react-moment";

import { QUERY_BILLING_INFO } from "ee/apollo/queries/billingInfo";
import { BillingInfo, BillingInfoVariables } from "ee/apollo/types/BillingInfo";
import { OrganizationByName_organizationByName_PrivateOrganization } from "apollo/types/OrganizationByName";
import VSpace from "components/VSpace";
import clsx from "clsx";
import { prettyPrintBytes } from "components/metrics/util";


const useStyles = makeStyles((theme) => ({
  paperPadding: {
    padding: theme.spacing(3)
  },
  billItem: {
    fontSize: theme.typography.pxToRem(36),
  },
  billTotal: {
    color: theme.palette.primary.dark,
  },
}));

export interface BillingInfoProps {
  organization: OrganizationByName_organizationByName_PrivateOrganization;
  setChangePlanDialog: (value: boolean) => void;
}

const ViewBillingPlan: FC<BillingInfoProps> = ({ organization, setChangePlanDialog }) => {
  const classes = useStyles();

  const { loading, error, data } = useQuery<BillingInfo, BillingInfoVariables>(QUERY_BILLING_INFO, {
    context: { ee: true },
    variables: {
      organizationID: organization.organizationID,
    },
  });

  if (error) {
    return <p>Error: {JSON.stringify(error)}</p>;
  }

  if (!data || !organization.readQuota || !organization.prepaidReadQuota || !organization.writeQuota || !organization.prepaidWriteQuota || !organization.scanQuota || !organization.prepaidScanQuota) {
    return <></>;
  }

  const billingInfo = data.billingInfo;
  const currencyFormatter = new Intl.NumberFormat('en-US', {style: 'currency', currency: 'USD'});

  return (
    <>
      <Typography variant="h2">
        Billing plan
      </Typography>
      <VSpace units={2}/>
      <Grid container direction="column">
        {/* Plan details */}
        <Grid item xs={12} md={8}>
          <Typography variant="h3" gutterBottom>
            Your current plan
          </Typography>
          <Paper className={classes.paperPadding} variant="outlined">
            <Grid container justify="space-between" alignItems="center">
              <Grid item>
                <Typography>
                  Plan name:
                </Typography>
              </Grid>
              <Grid item>
                <Typography variant="h4">
                  {billingInfo.billingPlan.description}
                </Typography>
              </Grid>
            </Grid>
            <VSpace units={1} />
            <Grid container justify="space-between">
              <Grid item>
                <Typography>
                  Current billing period:
                </Typography>
              </Grid>
              <Grid item>
                <Typography>
                  <Moment format="MMMM Do" subtract={{ days: 31 }}>{billingInfo.nextBillingTime}</Moment> to <Moment format="MMMM Do">{billingInfo.nextBillingTime}</Moment>
                </Typography>
              </Grid>
            </Grid>
            <VSpace units={3} />
            <Grid container justify="space-between">
              <Grid item>
              </Grid>
              <Grid item>
                <Grid container spacing={2}>
                  {!billingInfo.billingPlan.default && (
                    <Grid item>
                      <Button>Cancel</Button>
                    </Grid>
                  )}
                  <Grid item>
                    <Button variant="contained" onClick={() => setChangePlanDialog(true)}>
                      Upgrade plan
                    </Button>
                  </Grid>
                </Grid>
              </Grid>
            </Grid>
          </Paper>
        </Grid>

        <VSpace units={3} />
        {/* Your next bill */}
        <Grid item xs={12} md={8}>
          <Typography variant="h3" gutterBottom>
            Your next bill
          </Typography>
          <Paper className={classes.paperPadding} variant="outlined">
            <Grid container direction="column" alignItems="center">
              <Grid item>
                <Typography className={clsx(classes.billItem, classes.billTotal)}>
                  $0.00
                </Typography>
              </Grid>
              <Grid item>
                <Typography className={clsx(classes.billTotal)} variant="caption">
                  Total billed on <Moment format="MMMM Do">{billingInfo.nextBillingTime}</Moment>
                </Typography>
              </Grid>
            </Grid>
          </Paper>
        </Grid>

        <VSpace units={3} />
        {/* Your bill details */}
        <Grid item xs={12} md={8}>
          <Typography variant="h3" gutterBottom>
            Your bill details
          </Typography>
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
                <TableCell align="right">{currencyFormatter.format(billingInfo.billingPlan.readOveragePriceCents)}</TableCell>
                <TableCell align="right">
                  {currencyFormatter.format(
                    billingInfo.billingPlan.readOveragePriceCents *
                    (organization.readUsage - organization.prepaidReadQuota > 0 ? organization.readUsage - organization.prepaidReadQuota : 0)
                  )}
                </TableCell>
              </TableRow>
              <TableRow>
                <TableCell>Writes</TableCell>
                <TableCell align="right">{prettyPrintBytes(organization.prepaidWriteQuota)}</TableCell>
                <TableCell align="right">{prettyPrintBytes(organization.writeQuota - organization.prepaidWriteQuota)}</TableCell>
                <TableCell align="right">{prettyPrintBytes(organization.writeUsage)}</TableCell>
                <TableCell align="right">{currencyFormatter.format(billingInfo.billingPlan.writeOveragePriceCents)}</TableCell>
                <TableCell align="right">
                  {currencyFormatter.format(
                    billingInfo.billingPlan.writeOveragePriceCents *
                    (organization.writeUsage - organization.prepaidWriteQuota > 0 ? organization.writeUsage - organization.prepaidWriteQuota : 0)
                  )}
                </TableCell>
              </TableRow>
              <TableRow>
                <TableCell>Scans</TableCell>
                <TableCell align="right">{prettyPrintBytes(organization.prepaidScanQuota)}</TableCell>
                <TableCell align="right">{prettyPrintBytes(organization.scanQuota - organization.prepaidScanQuota)}</TableCell>
                <TableCell align="right">{prettyPrintBytes(organization.scanUsage)}</TableCell>
                <TableCell align="right">{currencyFormatter.format(billingInfo.billingPlan.scanOveragePriceCents)}</TableCell>
                <TableCell align="right">
                  {currencyFormatter.format(
                    billingInfo.billingPlan.scanOveragePriceCents *
                    (organization.scanUsage - organization.prepaidScanQuota > 0 ? organization.scanUsage - organization.prepaidScanQuota : 0)
                  )}
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </Grid>
      </Grid>
    </>
  );
};

export default ViewBillingPlan;
