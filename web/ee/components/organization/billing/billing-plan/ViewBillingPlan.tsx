import { useQuery } from "@apollo/client";
import _ from "lodash";
import {Grid, Typography } from "@material-ui/core";
import { makeStyles } from "@material-ui/core/styles";
import React, { FC } from "react";

import { OrganizationByName_organizationByName_PrivateOrganization } from "apollo/types/OrganizationByName";
import { QUERY_BILLING_INFO } from "ee/apollo/queries/billingInfo";
import { BillingInfo, BillingInfoVariables } from "ee/apollo/types/BillingInfo";
import ViewNextBillDetails from "./ViewNextBillDetails";
import ViewNextBillOverview from "./ViewNextBillOverview";
import clsx from "clsx";
import ViewCurrentPlan from "./ViewCurrentPlan";
import ContactUs from "../ContactUs";


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
  changePlan: (value: boolean) => void;
}

const ViewBillingPlan: FC<BillingInfoProps> = ({ organization, changePlan }) => {
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
      <Grid container spacing={6} alignItems="stretch">
        <Grid item xs={12} lg={8}>
          <Typography variant="h2" className={clsx(classes.sectionTitle, classes.firstSectionTitle)}>
            Your current plan
          </Typography>
          <ViewCurrentPlan billingInfo={billingInfo} changePlan={changePlan} />
        </Grid>

        <Grid item xs={12} lg={4}>
          <Typography variant="h2" className={clsx(classes.sectionTitle, classes.firstSectionTitle)}>
            Contact us
          </Typography>
          <ContactUs />
        </Grid>
      </Grid>

      <Typography variant="h2" className={classes.sectionTitle}>
        Overview of your next bill
      </Typography>
      <ViewNextBillOverview organization={organization} billingInfo={billingInfo} />

      <Typography variant="h2" className={classes.sectionTitle}>
        Details of your next bill
      </Typography>
      <ViewNextBillDetails organization={organization} billingInfo={billingInfo} />
    </>
  );
};

export default ViewBillingPlan;
