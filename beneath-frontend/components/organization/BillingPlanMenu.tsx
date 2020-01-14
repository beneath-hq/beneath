import React, { FC } from 'react';
import { Button } from "@material-ui/core";
import PaymentsByCard from "./PaymentsByCard"
import { ReactStripeElements } from 'react-stripe-elements';

const PRO_BILLING_PLAN_DESCRIPTION = "Professional"
const MONTHLY_BILLING_PLAN_STRING = "Monthly"

interface Props {
  stripe: ReactStripeElements.StripeProps | undefined
  organization_id: any
}

const BillingPlanMenu: FC<Props> = ({stripe, organization_id}) => {
  return (
    <div>
      <p>You are on the free plan. Consider the following plans: Pro plan (buy now), Enterprise plan (request demo)</p>
      <Button
        variant="contained"
        color="secondary"
        onClick={() => {
          return <PaymentsByCard stripe={stripe} organization_id={organization_id} billing_period={MONTHLY_BILLING_PLAN_STRING} description={PRO_BILLING_PLAN_DESCRIPTION}/>
        }}>
        Buy Now
      </Button>
      <Button
        variant="contained"
        color="secondary"
        onClick={() => {
          // TODO: fetch about.beneath.com/contact/demo
        }}>
        Request Demo
      </Button>
    </div>
  )
}

export default BillingPlanMenu;