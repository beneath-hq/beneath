import React, { FC, useEffect } from 'react';
import { CardElement } from 'react-stripe-elements';
import Loading from "../Loading";
import { TextField, Typography, Button } from "@material-ui/core";
import { useToken } from '../../hooks/useToken';
import { ReactStripeElements } from 'react-stripe-elements';
import connection from "../../lib/connection";

const PRO_BILLING_PLAN_ID = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb";

interface CardPaymentDetails {
  data: {
    organization_id: string,
    card: {
      Brand: string,
      Last4: string,
    },
    billing_details: {
      Name: string,
      Email: string,
      Phone: string,
      Address: {
        Line1: string,
        Line2: string,
        City: string,
        State: string,
        PostalCode: string,
        Country: string,
      }
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

interface CheckoutStateTypes {
  isBuyingNow: boolean,
  city: string,
  country: string,
  line1: string,
  line2: string,
  postal_code: string,
  state: string,
  email: string,
  firstname: string,
  lastname: string,
  phone: string,
  cardDetailsFormSubmit: number,
  stripeError: stripe.Error | undefined,
  paymentDetails: CardPaymentDetails | null,
  loading: boolean,
  status: stripe.setupIntents.SetupIntentStatus | null,
}

interface Props {
  stripe: ReactStripeElements.StripeProps | undefined;
  organization_id: any;
  billing_period: any;
  description: any;
}

const PaymentsByCard: FC<Props> = ({ stripe, organization_id, billing_period, description }) => {
  const [values, setValues] = React.useState<CheckoutStateTypes>({
    isBuyingNow: false,
    city: "",
    country: "",
    line1: "",
    line2: "",
    postal_code: "",
    state: "",
    email: "",
    firstname: "",
    lastname: "",
    phone: "",
    cardDetailsFormSubmit: 0,
    stripeError: undefined,
    paymentDetails: null,
    loading: false,
    status: null
  })
  const token = useToken();

  // Handle submission of Card Details Form
  const handleChange = (name: string) => (event: any) => {
    setValues({ ...values, [name]: event.target.value });
  };

  const handleCardDetailsFormSubmit = (ev: any) => {
    // We don't want to let default form submission happen here, which would refresh the page.
    ev.preventDefault();
    setValues({ ...values, ...{ cardDetailsFormSubmit: values.cardDetailsFormSubmit + 1 } })
    return
  }

  const CardBillingDetailsForm = (
    <div>
      <form onSubmit={handleCardDetailsFormSubmit}>
        <TextField
          id="firstname"
          label="First name"
          margin="normal"
          fullWidth
          required
          value={values.firstname}
          onChange={handleChange("firstname")}
        />
        <TextField
          id="lastname"
          label="Last name"
          margin="normal"
          fullWidth
          required
          value={values.lastname}
          onChange={handleChange("lastname")}
        />
        <TextField
          id="address_line1"
          label="Address line 1"
          margin="normal"
          fullWidth
          required
          value={values.line1}
          onChange={handleChange("line1")}
        />
        <TextField
          id="address_line2"
          label="Address line 2"
          margin="normal"
          fullWidth
          value={values.line2}
          onChange={handleChange("line2")}
        />
        <TextField
          id="address_city"
          label="City"
          margin="normal"
          fullWidth
          required
          value={values.city}
          onChange={handleChange("city")}
        />
        <TextField
          id="address_state"
          label="State"
          margin="normal"
          fullWidth
          required
          value={values.state}
          onChange={handleChange("state")}
        />
        <TextField
          id="address_zip"
          label="Zip code"
          margin="normal"
          fullWidth
          required
          value={values.postal_code}
          onChange={handleChange("postal_code")}
        />
        <TextField
          id="address_country"
          label="Country"
          margin="normal"
          helperText={!validateCountry(values.country) ? "Must be the two letter country code" : undefined}
          fullWidth
          required
          value={values.country}
          onChange={handleChange("country")}
        />
        <TextField
          id="email"
          label="Email address"
          margin="normal"
          fullWidth
          required
          value={values.email}
          onChange={handleChange("email")}
        />
        <TextField
          id="phone"
          label="Phone number"
          margin="normal"
          fullWidth
          required
          value={values.phone}
          onChange={handleChange("phone")}
        />
        <CardElement style={{ base: { fontSize: '18px', color: '#FFFFFF' } }} />
        <button>Submit</button>
      </form>

      {values.stripeError !== undefined && (
        <Typography variant="body1" color="error">
          {JSON.stringify(values.stripeError.message)}
        </Typography>
      )}
      {/* note: setup intent returns a "success" BEFORE we complete attaching the card to the customer in Stripe.
                so if we simulatenously trigger a refresh of the page, we'll see the old card info.
                so we need to wait a second before fetching the updated card info from stripe.   */}
      {values.status !== null && values.status === "succeeded" && (
        <Typography variant="body1">
          {values.status} -- need to refresh the page to see your card details on file
        </Typography>
      )}
      {values.status !== null && values.status !== "succeeded" && (
        <Typography variant="body1" color="error">
          {values.status}
        </Typography>
      )}
      <Button
        variant="contained"
        onClick={() => {
          setValues({ ...values, ...{ isBuyingNow: false } })
        }}>
        Back
      </Button>
    </div>
  )

  // When card form is submitted, initiate setupIntent
  const headers = { authorization: `Bearer ${token}` };
  let url = `${connection.API_URL}/billing/stripecard/generate_setup_intent`;
  url += `?organizationID=${organization_id}`;
  url += `&billingPlanID=${PRO_BILLING_PLAN_ID}`;

  useEffect(() => {
    let isMounted = true

    const fetchData = (async () => {
      if (!stripe) {
        return
      }

      setValues({ ...values, ...{ loading: true } })
      const res = await fetch(url, { headers });

      if (isMounted) {
        if (!res.ok) {
          setValues({ ...values, ...{ error: res.statusText } })
        }
        const intent: any = await res.json();

        const customerData: PaymentMethodData = {
          payment_method_data: {
            billing_details: {
              address: {
                city: values.city,
                country: values.country,
                line1: values.line1,
                line2: values.line2,
                postal_code: values.postal_code,
                state: values.state,
              },
              email: values.email,
              name: values.firstname + " " + values.lastname,
              phone: values.phone,
            }
          }
        }

        // handleCardSetup automatically pulls credit card info from the Card element
        // TODO from Stripe Docs: Note that stripe.handleCardSetup may take several seconds to complete. During that time, you should disable your form from being resubmitted and show a waiting indicator like a spinner. If you receive an error result, you should be sure to show that error to the customer, re-enable the form, and hide the waiting indicator.
        // ^ *** make sure not to "block" access to the Card Element by solely returning a loading spinner
        // TODO from Stripe Docs: Additionally, stripe.handleCardSetup may trigger a 3D Secure authentication challenge.This will be shown in a modal dialog and may be confusing for customers using assistive technologies like screen readers.You should make your form accessible by ensuring that success or error messages are clearly read out after this method completes
        const result: stripe.SetupIntentResponse = await stripe.handleCardSetup(intent.client_secret, customerData)
        if (result.error) {
          setValues({ ...values, ...{ stripeError: result.error, loading: false } })
        }
        if (result.setupIntent) {
          console.log(result.setupIntent) // TODO: if success, trigger refetch of billing info (to display current card on file), or just return a confirmation
          setValues({ ...values, ...{ stripeError: result.error, loading: false, status: result.setupIntent.status } })
        }
      }
    })

    fetchData()

    // avoid memory leak when component unmounts
    return () => {
      isMounted = false
    }
  }, [values.cardDetailsFormSubmit])

  // for paying customers, get current payment details 
  useEffect(() => {
    let isMounted = true

    const fetchData = async () => {
      console.log("FETCHING PAYMENT DETAILS")
      setValues({ ...values, ...{ loading: true } })
      let payment_details_url = `${connection.API_URL}/billing/stripecard/get_payment_details`;
      const res = await fetch(payment_details_url, { headers });

      if (isMounted) {
        if (!res.ok) {
          setValues({ ...values, ...{ error: res.statusText } })
        }

        const paymentDetails: CardPaymentDetails = await res.json();
        setValues({ ...values, ...{ paymentDetails: paymentDetails, loading: false } })
      }
    }

    fetchData()

    // avoid memory leak when component unmounts
    return () => {
      isMounted = false
    }
  }, [])

  if (values.loading) {
    return <Loading justify="center" />
  }

  if (values.paymentDetails == null) {
    return <p>Error: no payment details.. why?</p>;
  }

  // current card details
  const CardBillingDetails = (
      <div>
        <Typography variant="body1">Current billing plan: {description}</Typography>
        <Typography variant="body1">Current billing period: {billing_period}</Typography>
        <Typography variant="body1">Current billing details</Typography>
        <Typography variant="body1">Brand: {values.paymentDetails.data.card.Brand}</Typography>
        <Typography variant="body1">Last4: {values.paymentDetails.data.card.Last4}</Typography>
        <Typography variant="body1">Name: {values.paymentDetails.data.billing_details.Name}</Typography>
        <Typography variant="body1">Phone: {values.paymentDetails.data.billing_details.Phone}</Typography>
        <Typography variant="body1">Email: {values.paymentDetails.data.billing_details.Email}</Typography>
        <Typography variant="body1">Line1: {values.paymentDetails.data.billing_details.Address.Line1}</Typography>
        <Typography variant="body1">Line2: {values.paymentDetails.data.billing_details.Address.Line2}</Typography>
        <Typography variant="body1">City: {values.paymentDetails.data.billing_details.Address.City}</Typography>
        <Typography variant="body1">State: {values.paymentDetails.data.billing_details.Address.State}</Typography>
        <Typography variant="body1">PostalCode: {values.paymentDetails.data.billing_details.Address.PostalCode}</Typography>
        <Typography variant="body1">Country: {values.paymentDetails.data.billing_details.Address.Country}</Typography>
        <Button
          color="secondary"
          onClick={() => {
            setValues({ ...values, ...{ isBuyingNow: true } })
          }}>
          Change card on file
            </Button>
        <Typography variant="body1">Interested in upgrading to an Enterprise plan?</Typography>
        <Button
          variant="contained"
          color="secondary"
          onClick={() => {
            setValues({ ...values, ...{ isRequestingDemo: true } })
          }}>
          Request Demo
            </Button>
        <Typography variant="body1">contact us if you would like to discuss your billing plan</Typography>
      </div>
  )

  // either show input form or show current billing details
  if (values.isBuyingNow === true) {
    return CardBillingDetailsForm
  } else {
    return CardBillingDetails
  }
}

const validateCountry = (val: string) => {
  return val && val.length == 2;
};

export default PaymentsByCard