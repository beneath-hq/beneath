import React, { FC } from 'react';
import { useQuery } from "@apollo/react-hooks";
import Loading from "../../Loading";
import BillingPlanMenu from "./BillingPlanMenu"
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
  
  if (loading) {
    return <Loading justify="center" />;
  }

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

    // route to correct payment driver
    if (data.billingInfo.paymentsDriver === billing.STRIPECARD_DRIVER) {
      return <CardDetails billing_period={billingPeriod} description={data.billingInfo.billingPlan.description} billing_plan_id={data.billingInfo.billingPlan.billingPlanID}/>
    }

    if (data.billingInfo.paymentsDriver === billing.STRIPEWIRE_DRIVER) {
      return <WireDetails billing_period={billingPeriod} description={data.billingInfo.billingPlan.description} />
    }

    if (data.billingInfo.paymentsDriver === billing.ANARCHISM_DRIVER) {
      return <AnarchismDetails billing_period={billingPeriod} description={data.billingInfo.billingPlan.description} />
    }

    return <p> Error: payments driver is not supported. </p>
  }
};

export default ViewBilling;