import { Grid, makeStyles, Paper, Typography } from "@material-ui/core";
import { FC } from "react";
import Moment from "react-moment";

import { BillingInfo_billingInfo } from "ee/apollo/types/BillingInfo";
import clsx from "clsx";
import { OrganizationByName_organizationByName_PrivateOrganization } from "apollo/types/OrganizationByName";
import { QUERY_ORGANIZATION_MEMBERS } from "apollo/queries/organization";
import { useQuery } from "@apollo/client";
import { OrganizationMembers, OrganizationMembersVariables } from "apollo/types/OrganizationMembers";

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
  mathOperator: {
    fontSize: theme.typography.pxToRem(36)
  }
}));

interface Props {
  organization: OrganizationByName_organizationByName_PrivateOrganization;
  billingInfo: BillingInfo_billingInfo;
}

const ViewNextBillOverview: FC<Props> = ({organization, billingInfo}) => {
  const classes = useStyles();
  const currencyFormatter = new Intl.NumberFormat('en-US', {style: 'currency', currency: 'USD'});

  const {data, error, loading} = useQuery<OrganizationMembers, OrganizationMembersVariables>(QUERY_ORGANIZATION_MEMBERS, {
    variables: { organizationID: organization.organizationID },
  });

  if (!data || !organization.readQuota || !organization.prepaidReadQuota || !organization.writeQuota || !organization.prepaidWriteQuota || !organization.scanQuota || !organization.prepaidScanQuota) {
    return <></>;
  }

  const isOverageBased = billingInfo.billingPlan.readOveragePriceCents > 0 || billingInfo.billingPlan.writeOveragePriceCents > 0 || billingInfo.billingPlan.scanOveragePriceCents > 0;
  const readOverageTotal = billingInfo.billingPlan.readOveragePriceCents * (organization.readUsage - organization.prepaidReadQuota > 0 ? organization.readUsage - organization.prepaidReadQuota : 0);
  const writeOverageTotal = billingInfo.billingPlan.writeOveragePriceCents * (organization.writeUsage - organization.prepaidWriteQuota > 0 ? organization.writeUsage - organization.prepaidWriteQuota : 0);
  const scanOverageTotal = billingInfo.billingPlan.scanOveragePriceCents * (organization.scanUsage - organization.prepaidScanQuota > 0 ? organization.scanUsage - organization.prepaidScanQuota : 0);
  const overageTotal = readOverageTotal + writeOverageTotal + scanOverageTotal;

  const isSeatBased = billingInfo.billingPlan.seatPriceCents > 0;
  const numSeats = data.organizationMembers.filter((member) => member.billingOrganizationID === organization.organizationID).length;
  const seatTotal = billingInfo.billingPlan.seatPriceCents * numSeats;

  const billTotal = overageTotal + billingInfo.billingPlan.basePriceCents + seatTotal;

  return (
    <>
      <Paper className={classes.paperPadding} variant="outlined">
        <Grid container alignItems="center" spacing={3} justify="center">
          {/* BASE PRICE DETAIL */}
          <>
            {(isSeatBased || isOverageBased) && (
              <>
                <Grid item>
                  <Grid container direction="column" alignItems="center">
                    <Grid item>
                      <Typography className={classes.billItem}>
                        {currencyFormatter.format(billingInfo.billingPlan.basePriceCents / 100)}
                      </Typography>
                    </Grid>
                    <Grid item>
                      <Typography variant="caption">
                        Base price
                      </Typography>
                    </Grid>
                  </Grid>
                </Grid>
                <Grid item>
                  <Typography className={classes.mathOperator}>+</Typography>
                </Grid>
              </>
            )}
          </>

          {/* SEAT DETAIL */}
          <>
            {isSeatBased && (
              <>
                <Grid item>
                  <Grid container direction="column" alignItems="center">
                    <Grid item>
                      <Typography className={classes.billItem}>
                        {currencyFormatter.format(billingInfo.billingPlan.seatPriceCents / 100)}
                      </Typography>
                    </Grid>
                    <Grid item>
                      <Typography variant="caption">
                        Seat price
                      </Typography>
                    </Grid>
                  </Grid>
                </Grid>
                <Grid item>
                  <Typography className={classes.mathOperator}>x</Typography>
                </Grid>
                <Grid item>
                  <Grid container direction="column" alignItems="center">
                    <Grid item>
                      <Typography className={classes.billItem}>
                        {numSeats}
                      </Typography>
                    </Grid>
                    <Grid item>
                      <Typography variant="caption">
                        Number of seats
                      </Typography>
                    </Grid>
                  </Grid>
                </Grid>
                <Grid item>
                  <Typography className={classes.mathOperator}>
                    {isOverageBased ? "+" : "="}
                  </Typography>
                </Grid>
              </>
            )}
          </>

          {/* OVERAGE DETAIL */}
          <>
          {isOverageBased && (
            <>
            <Grid item>
              <Grid container direction="column" alignItems="center">
                <Grid item>
                  <Typography className={clsx(classes.billItem)}>
                    {currencyFormatter.format(overageTotal / 100)}
                  </Typography>
                </Grid>
                <Grid item>
                  <Typography variant="caption">
                    Overage
                  </Typography>
                </Grid>
              </Grid>
            </Grid>
            <Grid item>
              <Typography className={classes.mathOperator}>=</Typography>
            </Grid>
            </>
          )}
          </>

          {/* TOTAL */}
          <Grid item>
            <Grid container direction="column" alignItems="center" justify="center">
              <Grid item>
                <Typography className={clsx(classes.billItem, classes.billTotal)}>
                  {currencyFormatter.format(billTotal / 100)}
                </Typography>
              </Grid>
              <Grid item xs={12}>
                <Typography className={clsx(classes.billTotal)} variant="caption" align="center">
                  Total billed on <Moment format="MMMM Do">{billingInfo.nextBillingTime}</Moment>
                </Typography>
              </Grid>
            </Grid>
          </Grid>
        </Grid>
      </Paper>
    </>
  );
};

export default ViewNextBillOverview;