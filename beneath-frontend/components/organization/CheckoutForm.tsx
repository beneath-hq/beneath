// Note: very much a WIP!

import React, { FC, useEffect } from 'react';
import { useQuery } from "@apollo/react-hooks";
import { injectStripe, ReactStripeElements } from 'react-stripe-elements';
import { CardElement } from 'react-stripe-elements';
import { OrganizationByName_organizationByName } from "../../apollo/types/OrganizationByName";
import { makeStyles, TextField, Typography } from "@material-ui/core";
import Loading from "../Loading";
import connection from "../../lib/connection";
import { useToken } from '../../hooks/useToken';
import organization from '../../pages/organization';
import { BillingInfo, BillingInfoVariables } from '../../apollo/types/BillingInfo';
import { QUERY_BILLING_INFO } from '../../apollo/queries/bililnginfo';

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

interface CardPaymentDetails {
  data: {
    organization_id: string,
    card: {
      Brand: string,
      Last4: string,
    }
  },
  error: string | undefined
}

interface PaymentMethodData {
  payment_method_data: {
    billing_details: {
      address: {
        city: string,
        country: string,
        line1: string,
        line2: string,
        postal_code: string,
        state: string
      },
      email: string,
      name: string,
      phone: string
    }
  }
}

interface Props {
  stripe: ReactStripeElements.StripeProps | undefined;
  organization: OrganizationByName_organizationByName;
}

interface CheckoutStateTypes {
  customerData: PaymentMethodData | null,
  isLoading: boolean,
  error: string | null,
  stripeError: stripe.Error | undefined,
  isReady: boolean,
  stopInfinite: boolean,
  paymentDetails: CardPaymentDetails | null, // WirePaymentDetails | AnarchismPaymentDetails (?) | null, // TODO: this will depend on the driver
  isIntentLoading: boolean,
  status: stripe.setupIntents.SetupIntentStatus | null,
}

const CheckoutForm: FC<Props> = ({ stripe, organization }) => {
  const [values, setValues] = React.useState<CheckoutStateTypes>({
    customerData: null,
    isLoading: false,
    error: null,
    stripeError: undefined,
    isReady: false,
    stopInfinite: true,
    paymentDetails: null,
    isIntentLoading: false,
    status: null,
  })
  const classes = useStyles();
  const token = useToken();
  const PRO_BILLING_PLAN_ID = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb";

  // 0. fetch organization -> billing info -> payment driver
  const { loading, error, data } = useQuery<BillingInfo, BillingInfoVariables>(QUERY_BILLING_INFO, {
    variables: {
      organizationID: organization.organizationID,
    },
  });

  if (loading) {
    return <Loading justify="center" />;
  }

  if (error) {
    return <p>Error: {JSON.stringify(error)}</p>;
  }

  if (!data) {
    return <p>No error and no billing info... why?</p> 
  }

  if (data.billingInfo.paymentsDriver === "stripecard") {
    // 1. Get current payment details upon page load
    let payment_details_url = `${connection.API_URL}/billing/stripecard/get_payment_details`; // TODO: make "stripecard" dynamic based on driver
    const headers = { authorization: `Bearer ${token}` };

    // TODO: try to do the call only once, by checking the Stripe props (which is what changes from the first mount to the second); see BillingTab.tsx
    // console.log(typeof stripe)
    // console.log(Object.prototype.toString.call(stripe))
    if ((true) && values.stopInfinite) { // typeof stripe === something
      console.log("GOT HERE")
      setValues({ ...values, ...{ isReady: true, stopInfinite: false } })
    }

    useEffect(() => {
      (async () => {
        if (values.isReady) {
          console.log("FETCHING PAYMENT DETAILS")
          setValues({ ...values, ...{ isLoading: true } })
          const res = await fetch(payment_details_url, { headers });
          if (!res.ok) {
            setValues({ ...values, ...{ error: res.statusText } })
          }
          const details: CardPaymentDetails = await res.json();
          setValues({ ...values, ...{ paymentDetails: details, isLoading: false } })
        }
      })()
    }, [values.isReady])

    // 2. Handle submission of Card details and Customer Data
    const handleSubmit = (ev: any) => {
      // We don't want to let default form submission happen here, which would refresh the page.
      ev.preventDefault();

      // TODO: get customerData from Form below; validate the customerData before setting state; handle errors
      setValues({
        ...values, ...{
          customerData: {
            payment_method_data: {
              billing_details: {
                address: {
                  city: "Boston",
                  country: "US",
                  line1: "74 Stone Rd",
                  line2: "none",
                  postal_code: "02478",
                  state: "MA"
                },
                email: "ericpgreen2@gmail.com",
                name: "Eric Green",
                phone: "6177101732",
              }
            }
          }
        }
      })

      return
    };

    // 3. When card form is submitted (and Customer Data changes), initiate setupIntent
    let url = `${connection.API_URL}/billing/stripecard/generate_setup_intent`;
    url += `?organizationID=${organization.organizationID}`;
    url += `&billingPlanID=${PRO_BILLING_PLAN_ID}`;

    useEffect(() => {
      (async () => {
        console.log("into setup intent use effect")
        if (!stripe || !values.customerData) {
          return;
        }
        console.log("got here")
        setValues({ ...values, ...{ isIntentLoading: true } })

        const res = await fetch(url, { headers });
        if (!res.ok) {
          setValues({ ...values, ...{ error: res.statusText } })
        }
        const intent: any = await res.json();

        // handleCardSetup automatically pulls credit card info from the Card element
        // TODO from Stripe Docs: Note that stripe.handleCardSetup may take several seconds to complete. During that time, you should disable your form from being resubmitted and show a waiting indicator like a spinner. If you receive an error result, you should be sure to show that error to the customer, re-enable the form, and hide the waiting indicator.
        // ^ *** make sure not to "block" access to the Card Element by solely returning a loading spinner
        // TODO from Stripe Docs: Additionally, stripe.handleCardSetup may trigger a 3D Secure authentication challenge.This will be shown in a modal dialog and may be confusing for customers using assistive technologies like screen readers.You should make your form accessible by ensuring that success or error messages are clearly read out after this method completes
        const result: stripe.SetupIntentResponse = await stripe.handleCardSetup(intent.client_secret, values.customerData)
        if (result.error) {
          setValues({ ...values, ...{ stripeError: result.error, isIntentLoading: false } })
        }
        if (result.setupIntent) {
          console.log(result.setupIntent)
          setValues({ ...values, ...{ stripeError: result.error, isIntentLoading: false, status: result.setupIntent.status } })
        }
      })();
    }, [values.customerData])

    // 4. handle states
    if (values.isLoading || !values.paymentDetails) {
      return <Loading justify="center" />;
    }

    if (!values.paymentDetails.data) {
      return <p>No existing payment details.</p>;
    }

    if (values.error) {
      return <p>Error: {JSON.stringify(values.error)}</p>;
    }

    return (
      <div>
        <Typography variant="body1">Current billing details</Typography>
        <Typography variant="body1">Brand: {values.paymentDetails.data.card.Brand}</Typography>
        <Typography variant="body1">Last4: {values.paymentDetails.data.card.Last4}</Typography>
        <form > 
          {/* onSubmit={handleSubmit} */}
          {/* <TextField
            id="firstname"
            label="Firstname"
            margin="normal"
            fullWidth
            required
          />
          <TextField
            id="lastname"
            label="Lastname"
            margin="normal"
            fullWidth
            required
          />
          <TextField
            id="address_line1"
            label="Address_line1"
            margin="normal"
            fullWidth
            required
          />
          <TextField
            id="address_line2"
            label="Address_line2"
            margin="normal"
            fullWidth
          />
          <TextField
            id="address_city"
            label="Address_city"
            margin="normal"
            fullWidth
            required
          />
          <TextField
            id="address_state"
            label="Address_state"
            margin="normal"
            fullWidth
            required
          />
          <TextField
            id="address_zip"
            label="Address_zip"
            margin="normal"
            fullWidth
            required
          />
          <TextField
            id="address_country"
            label="Address_country"
            margin="normal"
            fullWidth
            required
          /> */}
          <CardElement style={{ base: { fontSize: '18px', color: '#FFFFFF' } }} />
          <button>Submit</button>
          { values.stripeError !== undefined && (
            <Typography variant="body1" color="error">
              {JSON.stringify(values.stripeError)}
            </Typography>
          )}
          {status !== null && (
            <Typography variant="body1" color="error">
              { status }
            </Typography>
          )}
        </form>
      </div>
    );
  }

  if (data.billingInfo.paymentsDriver === "stripewire") {
    return <p> You're paying by wire </p> 
  }

  if (data.billingInfo.paymentsDriver === "anarchism") {
    return <p> User is on the Free plan. To upgrade, need to pick a billing plan and payment driver. </p>
  }

  if (!(["stripecard", "stripewire", "anarchsim"].includes(data.billingInfo.paymentsDriver))) {
    return <p> No error and payments driver is not supported. Hmm... </p>
  }
  
  return <p> Unexplained behavior </p> 
};
