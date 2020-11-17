import { useQuery } from "@apollo/client";
import _ from "lodash";
import { Button, Grid, Paper, Typography } from "@material-ui/core";
import { makeStyles } from "@material-ui/core/styles";
import React, { FC } from "react";
import Moment from "react-moment";

import { OrganizationByName_organizationByName_PrivateOrganization } from "apollo/types/OrganizationByName";
import { QUERY_BILLING_INFO } from "ee/apollo/queries/billingInfo";
import { BillingInfo, BillingInfoVariables } from "ee/apollo/types/BillingInfo";
import VSpace from "components/VSpace";
import ViewNextBillDetails from "./ViewNextBillDetails";
import ViewNextBillOverview from "./ViewNextBillOverview";
import clsx from "clsx";


const useStyles = makeStyles((theme) => ({
  paperPadding: {
    padding: theme.spacing(3)
  },
  sectionTitle: {
    marginTop: theme.spacing(6),
    marginBottom: theme.spacing(3)
  },
  firstSectionTitle: {
    marginTop: theme.spacing(0),
  }
}));

export interface BillingInfoProps {
  organization: OrganizationByName_organizationByName_PrivateOrganization;
  cancelPlan: (value: boolean) => void;
  changePlan: (value: boolean) => void;
}

const ViewBillingPlan: FC<BillingInfoProps> = ({ organization, cancelPlan, changePlan }) => {
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

  return (
    <>
      <Grid container direction="column">
        <Grid item xs={12} md={8}>
          <Typography variant="h3" className={clsx(classes.sectionTitle, classes.firstSectionTitle)}>
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
                      <Button onClick={() => cancelPlan(true)}>Cancel plan</Button>
                    </Grid>
                  )}
                  <Grid item>
                    <Button variant="contained" onClick={() => changePlan(true)}>
                      Upgrade plan
                    </Button>
                  </Grid>
                </Grid>
              </Grid>
            </Grid>
          </Paper>
        </Grid>

        <Grid item xs={12} md={8}>
          <Typography variant="h3" className={classes.sectionTitle}>
            Overview of your next bill
          </Typography>
          <ViewNextBillOverview organization={organization} billingInfo={billingInfo} />
        </Grid>

        <Grid item xs={12} md={8}>
          <Typography variant="h3" className={classes.sectionTitle}>
            Details of your next bill
          </Typography>
          <ViewNextBillDetails organization={organization} billingInfo={billingInfo} />
        </Grid>
      </Grid>
    </>
  );
};

export default ViewBillingPlan;
