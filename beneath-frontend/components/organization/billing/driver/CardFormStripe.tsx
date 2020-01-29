import React from "react";
import { StripeProvider, Elements } from 'react-stripe-elements';
import CardForm from "./CardForm";

interface ViewCardFormProps {
  billing_plan_id: string;
}

interface ViewCardFormState {
  stripe: stripe.Stripe | null;
  billing_plan_id: string;
}

class CardFormStripe extends React.Component<ViewCardFormProps, ViewCardFormState> {
  constructor(props: ViewCardFormProps) {
    super(props);
    this.state = { stripe: null, billing_plan_id: props.billing_plan_id};
  }

  componentDidMount() {
    // Create Stripe instance in componentDidMount (componentDidMount only fires in browser/DOM environment) 
    // note that updating the state like this will cause the CardForm to fire/initially render twice
    this.setState({ stripe: window.Stripe('pk_test_L140lbWnkGmtqSiw8rH2wcNs00otQFgbbr') })
  }

  render() {
    return (
      <StripeProvider stripe={this.state.stripe}>
        <Elements>
          <CardForm billing_plan_id={this.state.billing_plan_id}/>
        </Elements>
      </StripeProvider>
    );
  }
}

export default CardFormStripe;
