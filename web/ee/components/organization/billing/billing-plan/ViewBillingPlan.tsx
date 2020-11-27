import { useQuery } from "@apollo/client";
import _ from "lodash";
import {Paper, Typography } from "@material-ui/core";
import { makeStyles } from "@material-ui/core/styles";
import React, { FC } from "react";

import { OrganizationByName_organizationByName_PrivateOrganization } from "apollo/types/OrganizationByName";
import VSpace from "components/VSpace";
import { QUERY_BILLING_INFO } from "ee/apollo/queries/billingInfo";
import { BillingInfo, BillingInfoVariables } from "ee/apollo/types/BillingInfo";
import ViewNextBillDetails from "./ViewNextBillDetails";
import ViewNextBillOverview from "./ViewNextBillOverview";
import ViewCurrentPlan from "./ViewCurrentPlan";
import Actions from "./Actions";


const useStyles = makeStyles((theme) => ({
  paper: {
    padding: theme.spacing(3)
  },
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
        <Typography variant="h1" gutterBottom>
          Billing plan
        </Typography>
        <Typography variant="body2" color="textSecondary">
          Your current billing plan and information about your next payment
        </Typography>

        <VSpace units={3} />
        <ViewCurrentPlan billingInfo={billingInfo} />
        <VSpace units={3} />
        <Actions billingInfo={billingInfo} organization={organization} />
        <VSpace units={3} />
        <ViewNextBillOverview organization={organization} billingInfo={billingInfo} />
        <VSpace units={3} />
        <ViewNextBillDetails organization={organization} billingInfo={billingInfo} />
      </Paper>
    </>
  );
};

export default ViewBillingPlan;
