import React, { FC } from 'react';
import { useQuery } from "@apollo/react-hooks";
import { injectStripe, ReactStripeElements } from 'react-stripe-elements';
import { OrganizationByName_organizationByName } from "../../apollo/types/OrganizationByName";
import { makeStyles } from "@material-ui/core";
import Loading from "../Loading";
import BillingPlanMenu from "./BillingPlanMenu"
import PaymentsByCard from "./PaymentsByCard"
import PaymentsByWire from "./PaymentsByWire"
import PaymentsByAnarchism from "./PaymentsByAnarchism"
import { BillingInfo, BillingInfoVariables } from '../../apollo/types/BillingInfo';
import { QUERY_BILLING_INFO } from '../../apollo/queries/bililnginfo';

const FREE_BILLING_PLAN_DESCRIPTION = "FREE"
const MONTHLY_BILLING_PLAN_STRING = "monthly"

const useStyles = makeStyles((theme) => ({
  noDataCaption: {
    color: theme.palette.text.secondary,
  },
}));

interface CheckoutParams {
  organization: OrganizationByName_organizationByName
}

class CheckoutFormWrapper extends React.Component<ReactStripeElements.InjectedStripeProps & CheckoutParams, CheckoutParams> {
  constructor(props: ReactStripeElements.InjectedStripeProps & CheckoutParams) {
    super(props);
    this.state = { organization: props.organization };
  }

  render() {
    return <CheckoutForm stripe={this.props.stripe} organization={this.state.organization} />
  }
}

export default injectStripe(CheckoutFormWrapper);

interface Props {
  stripe: ReactStripeElements.StripeProps | undefined;
  organization: OrganizationByName_organizationByName;
}

const CheckoutForm: FC<Props> = ({ stripe, organization }) => {  
  // fetch organization's billing info for the billing plan and payment driver
  const { loading, error, data } = useQuery<BillingInfo, BillingInfoVariables>(QUERY_BILLING_INFO, {
    variables: {
      organizationID: organization.organizationID,
    },
  });
  
  if (loading) {
    return <Loading justify="center" />;
  }

  if (error || !data) {
    return <p>Error: {JSON.stringify(error)}</p>;
  }

  // if not yet a paying customer, route to billing plan menu, else route to current payment info
  if (data.billingInfo.billingPlan.description === FREE_BILLING_PLAN_DESCRIPTION) {
    return <BillingPlanMenu stripe={stripe} organization_id={organization.organizationID} />
  } else {
    // decode billing period
    if (data.billingInfo.billingPlan.period === '\u0003') {
      var billingPeriod: string = MONTHLY_BILLING_PLAN_STRING
    } else {
      return <p>Error: your organization has an unknown billing plan period</p>;
    }

    // route to correct payment driver
    if (data.billingInfo.paymentsDriver === "stripecard") {
      return <PaymentsByCard stripe={stripe} organization_id={organization.organizationID} billing_period={billingPeriod} description={data.billingInfo.billingPlan.description} />
    }

    if (data.billingInfo.paymentsDriver === "stripewire") {
      return <PaymentsByWire billing_period={billingPeriod} description={data.billingInfo.billingPlan.description} />
    }

    if (data.billingInfo.paymentsDriver === "anarchism") {
      return <PaymentsByAnarchism billing_period={billingPeriod} description={data.billingInfo.billingPlan.description} />
    }

    return <p> Error: payments driver is not supported. </p>
  }
};
