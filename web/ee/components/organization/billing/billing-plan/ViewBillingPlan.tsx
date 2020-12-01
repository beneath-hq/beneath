import { useQuery } from "@apollo/client";
import _ from "lodash";
import {Paper, Typography } from "@material-ui/core";
import { makeStyles } from "@material-ui/core/styles";
import React, { FC } from "react";

import { OrganizationByName_organizationByName_PrivateOrganization } from "apollo/types/OrganizationByName";
import { QUERY_BILLING_INFO } from "ee/apollo/queries/billingInfo";
import { BillingInfo, BillingInfoVariables } from "ee/apollo/types/BillingInfo";
import ViewNextBillDetails from "./ViewNextBillDetails";
import ViewNextBillOverview from "./ViewNextBillOverview";
import ViewCurrentPlan from "./ViewCurrentPlan";

const useStyles = makeStyles((theme) => ({
  paper: {
    padding: theme.spacing(3)
  },
  paperTitle: {
    marginBottom: theme.spacing(1),
  }
}));

export interface BillingInfoProps {
  organization: OrganizationByName_organizationByName_PrivateOrganization;
}

const ViewBillingPlan: FC<BillingInfoProps> = ({ organization }) => {
  const classes = useStyles();

  const { loading, error, data } = useQuery<BillingInfo, BillingInfoVariables>(QUERY_BILLING_INFO, {
    context: { ee: true },
    variables: {
      organizationID: organization.organizationID,
    },
  });

  if (!data || error || !organization.readQuota || !organization.prepaidReadQuota || !organization.writeQuota || !organization.prepaidWriteQuota || !organization.scanQuota || !organization.prepaidScanQuota) {
    return null;
  }

  const billingInfo = data.billingInfo;

  return (
    <>
      <Paper variant="outlined" className={classes.paper}>
        <Typography variant="h1" className={classes.paperTitle}>
          Billing plan
        </Typography>
        <Typography variant="body2" color="textSecondary">
          Your current billing plan and information about your next payment
        </Typography>

        <ViewCurrentPlan organization={organization} billingInfo={billingInfo} />
        <ViewNextBillOverview organization={organization} billingInfo={billingInfo} />
        <ViewNextBillDetails organization={organization} billingInfo={billingInfo} />
      </Paper>
    </>
  );
};

export default ViewBillingPlan;
