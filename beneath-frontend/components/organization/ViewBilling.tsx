import React from "react";
import {
  StripeProvider,
  Elements,
} from 'react-stripe-elements';

import { OrganizationByName_organizationByName } from "../../apollo/types/OrganizationByName";
import CheckoutForm from "./CheckoutForm";

interface ViewBillingProps {
  organization: OrganizationByName_organizationByName;
}

interface ViewBillingState {
  stripe: stripe.Stripe | null;
  organization: OrganizationByName_organizationByName;
}

class ViewBilling extends React.Component<ViewBillingProps, ViewBillingState> {
  constructor(props: ViewBillingProps) {
    super(props);
    this.state = { stripe: null, organization: props.organization};
  }

  componentDidMount() {
    // Create Stripe instance in componentDidMount (componentDidMount only fires in browser/DOM environment) 
    // note that updating the state like this will cause the CheckoutForm to fire/initially render twice
    this.setState({ stripe: window.Stripe('pk_test_L140lbWnkGmtqSiw8rH2wcNs00otQFgbbr') })
  }

  render() {
    return (
      <StripeProvider stripe={this.state.stripe}>
        <Elements>
          <CheckoutForm organization={this.state.organization}/>
        </Elements>
      </StripeProvider>
    );
  }
}

export default ViewBilling;
