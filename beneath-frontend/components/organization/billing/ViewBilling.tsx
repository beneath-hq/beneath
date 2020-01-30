import React, { FC } from 'react';
import { Grid } from "@material-ui/core"
import { useQuery } from "@apollo/react-hooks";
import BillingPlanMenu from "./BillingPlanMenu"
import CurrentBillingPlan from './CurrentBillingPlan'
import CardDetails from "./driver/CardDetails"
import WireDetails from "./driver/WireDetails"
import AnarchismDetails from "./driver/AnarchismDetails"
import { BillingInfo, BillingInfoVariables } from '../../../apollo/types/BillingInfo';
import { QUERY_BILLING_INFO } from '../../../apollo/queries/bililnginfo';
import billing from "../../../lib/billing"
import useMe from "../../../hooks/useMe";

const ViewBilling: FC = () => {  
  const me = useMe(); // Q: is this in apollo local state?
  if (!me) {
    return <p>Need to log in to see your current billing plan</p>
  }

  // fetch organization's billing info for the billing plan and payment driver
  const { loading, error, data } = useQuery<BillingInfo, BillingInfoVariables>(QUERY_BILLING_INFO, {
    variables: {
      organizationID: me.organization.organizationID,
    },
  });

  if (error || !data) {
    return <p>Error: {JSON.stringify(error)}</p>;
  }

  // if not yet a paying customer, route to billing plan menu
  // if already a paying customer, route to current payment info
  if (data.billingInfo.billingPlan.description === billing.FREE_BILLING_PLAN_DESCRIPTION) {
    return <BillingPlanMenu/>
  } else {
    // decode billing period
    if (data.billingInfo.billingPlan.period === '\u0003') {
      var billingPeriod: string = billing.MONTHLY_BILLING_PLAN_STRING
    } else {
      return <p>Error: your organization has an unknown billing plan period</p>;
    }

    // get payment details for correct driver
    if (data.billingInfo.paymentsDriver === billing.STRIPECARD_DRIVER) {
      return (
        // CurrentBillingPlan is bundled inside CardDetails so that CurrentBillingPlan doesn't persist when editing someone edits their card details
        <CardDetails billing_plan_id={data.billingInfo.billingPlan.billingPlanID} billingPeriod={billingPeriod} description={data.billingInfo.billingPlan.description} />
      )
    } else if (data.billingInfo.paymentsDriver === billing.STRIPEWIRE_DRIVER) {
      return (
        <Grid container spacing={2}>
          <CurrentBillingPlan billing_period={billingPeriod} description={data.billingInfo.billingPlan.description} />
          <WireDetails />
        </Grid>
      )
    } else if (data.billingInfo.paymentsDriver === billing.ANARCHISM_DRIVER) {
      return (
        <Grid container spacing={2}>
          <CurrentBillingPlan billing_period={billingPeriod} description={data.billingInfo.billingPlan.description} />
          <AnarchismDetails />
        </Grid>
      )
    } else {
      return (
        <Grid container spacing={2}>
          <CurrentBillingPlan billing_period={billingPeriod} description={data.billingInfo.billingPlan.description} />
          <p> Error: payments driver is not supported. </p>
        </Grid>
      )
    }
  }
};

export default ViewBilling;