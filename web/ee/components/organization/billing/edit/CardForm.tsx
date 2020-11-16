import { Grid, Typography} from "@material-ui/core";
import _ from "lodash";
import React, { FC } from "react";
import { CardElement, Elements, useElements, useStripe } from "@stripe/react-stripe-js";
import { loadStripe } from "@stripe/stripe-js";

import { OrganizationByName_organizationByName_PrivateOrganization } from "apollo/types/OrganizationByName";
import { Formik, Form, Field } from "formik";
import FormikTextField from "components/formik/TextField";
import FormikSelectField from "components/formik/SelectField";
import SubmitControl from "components/forms/SubmitControl";
import { COUNTRY_CODES, STRIPE_KEY } from "ee/lib/billing";
import { API_URL } from "lib/connection";
import { useToken } from "hooks/useToken";
import useMe from "hooks/useMe";

interface Props {
  organization: OrganizationByName_organizationByName_PrivateOrganization;
  closeDialog: (value: boolean) => void;
}

interface Country {
  label: string;
  code: string;
}

const CardFormElement: FC<Props> = ({ organization, closeDialog }) => {
  const token = useToken();
  const me = useMe();
  const stripe = useStripe();
  const elements = useElements();

  const initialValues = {
    cardholder: "",
    line1: "",
    line2: "",
    city: "",
    state: "",
    postalCode: "",
    country: null as (null | Country)
  };

  return (
    <Formik
      initialValues={initialValues}
      onSubmit={async (values, actions) => {
        const headers = { authorization: `Bearer ${token}` };
        let url = `${API_URL}/ee/billing/stripecard/generate_setup_intent`;
        url += `?organizationID=${organization.organizationID}`;
        const res = await fetch(url, { headers });
        const intent: any = await res.json();
        if (!res.ok) {
          actions.setStatus(intent.error);
          return;
        }

        // handleCardSetup automatically pulls credit card info from the Card element
        // TODO from Stripe Docs: stripe.handleCardSetup may trigger a 3D Secure authentication challenge.
        // This will be shown in a modal dialog and may be confusing for customers using
        // assistive technologies like screen readers. You should make your form accessible by ensuring
        // that success or error messages are clearly read out after this method completes
        const cardElement = elements?.getElement(CardElement);
        if (!cardElement) {
          actions.setStatus(intent.error);
          return;
        }

        const response: stripe.SetupIntentResponse | any =
          await stripe?.confirmCardSetup(intent.client_secret, {
            payment_method: {
              card: cardElement,
              billing_details: {
                address: {
                  city: values.city,
                  country: values.country?.code,
                  line1: values.line1,
                  line2: values.line2,
                  postal_code: values.postalCode,
                  state: values.state,
                },
                email: me?.personalUser?.email, // Stripe receipts will be sent to the user's Beneath email address
                name: values.cardholder,
              }
            }
          });

        if (response.error) {
          actions.setStatus(response.error.message);
        }
        if (response.setupIntent) {
          actions.setStatus(response.setupIntent.status);
        }

        // TODO:
        // 1. when success, close the dialog and add an alert to the ViewBilling component "successful addition of card"
        // trigger a refresh of the billingMethods() query

        // NOTES:
        // - use actions.setFieldError() or actions.setStatus() (a global error or success)
        // - run onCompleted() (which should close the dialog and show a success message)

        return;
      }}
    >
      {({ isSubmitting, status }) => (
        <Form title="Add a credit card">
          {/* <Typography component="h2" variant="h1" gutterBottom>
            Add a credit card
          </Typography> */}
          {/* <Typography component="h2" variant="h1" gutterBottom>
            Billing information
          </Typography> */}
          <Grid container justify="space-between" spacing={2}>
            <Grid item xs={12} md={6}>
              <Field
                name="cardholder"
                validate={(val: string) => {
                  if (!val) return "Required field";
                }}
                component={FormikTextField}
                label="Name on card"
                required
              />
            </Grid>
          </Grid>
          <Grid container justify="space-between" spacing={2}>
            <Grid item xs={12} md={6}>
              <Field
                name="line1"
                validate={(val: string) => {
                  if (!val) return "Required field";
                }}
                component={FormikTextField}
                label="Address line 1"
                required
              />
            </Grid>
            <Grid item xs={12} md={6}>
              <Field
                name="line2"
                component={FormikTextField}
                label="Address line 2"
              />
            </Grid>
          </Grid>
          <Grid container justify="space-between" spacing={2}>
            <Grid item xs={12} md={6}>
              <Field
                name="city"
                validate={(val: string) => {
                  if (!val) return "Required field";
                }}
                component={FormikTextField}
                label="City"
                required
              />
            </Grid>
            <Grid item xs={12} md={6}>
              <Field
                name="state"
                validate={(val: string) => {
                  if (!val) return "Required field";
                }}                component={FormikTextField}
                label="State/Province/Region"
                required
              />
            </Grid>
          </Grid>
          <Grid container justify="space-between" spacing={2}>
            <Grid item xs={12} md={6}>
              <Field
                name="postalCode"
                validate={(val: string) => {
                  if (!val) return "Required field";
                }}
                component={FormikTextField}
                label="Zip / Postal code"
                required
              />
            </Grid>
            <Grid item xs={12} md={6}>
              <Field
                name="country"
                component={FormikSelectField}
                label="Country"
                required
                options={COUNTRY_CODES}
                getOptionLabel={(option: Country) => option.label}
                getOptionSelected={(option: Country, value: Country) => {
                  return option.code === value.code;
                }}
              />
            </Grid>
          </Grid>
          <Typography component="h2" variant="h1" gutterBottom>
            Card details
          </Typography>
          <CardElement
            options={{
              style: {
                base: { fontSize: "18px", color: "#FFFFFF" }
              }
            }} />
          <SubmitControl label="Submit" errorAlert={status} disabled={isSubmitting} />
        </Form>
      )}
    </Formik>
  );
};

const stripePromise = loadStripe(STRIPE_KEY);

const CardForm: FC<Props> = ({ organization, closeDialog }) => {
  return (
    <Elements stripe={stripePromise}>
      <CardFormElement organization={organization} closeDialog={closeDialog} />
    </Elements>
  );
};

export default CardForm;
