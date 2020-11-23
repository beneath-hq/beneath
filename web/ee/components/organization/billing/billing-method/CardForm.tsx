import { useApolloClient } from "@apollo/client";
import { Grid, makeStyles, Typography} from "@material-ui/core";
import _ from "lodash";
import React, { FC } from "react";
import { CardElement, Elements, useElements, useStripe } from "@stripe/react-stripe-js";
import { loadStripe } from "@stripe/stripe-js";

import { OrganizationByName_organizationByName_PrivateOrganization } from "apollo/types/OrganizationByName";
import { Formik, Form } from "formik";
import SubmitControl from "components/forms/SubmitControl";
import { STRIPE_KEY } from "ee/lib/billing";
import { API_URL } from "lib/connection";
import { useToken } from "hooks/useToken";
import useMe from "hooks/useMe";
import Loading from "components/Loading";
import VSpace from "components/VSpace";

const useStyles = makeStyles((theme) => ({
  sectionTitle: {
    marginTop: theme.spacing(4),
    marginBottom: theme.spacing(3)
  },
  cardInput: {
    marginBottom: theme.spacing(3)
  },
  loadingText: {
    marginTop: theme.spacing(3),
    fontWeight: "bold"
  }
}));

interface Props {
  organization: OrganizationByName_organizationByName_PrivateOrganization;
  openDialogFn: (value: boolean) => void;
}

const CardFormElement: FC<Props> = ({ organization, openDialogFn }) => {
  const classes = useStyles();
  const token = useToken();
  const me = useMe();
  const stripe = useStripe();
  const elements = useElements();
  const client = useApolloClient();
  const [loading, setLoading] = React.useState(false);

  if (loading) {
    return (
      <>
        <VSpace units={10}/>
        <Grid container direction="column" alignItems="center">
          <Grid item>
            <Loading justify="center" size={30}/>
          </Grid>
          <Grid item>
            <Typography className={classes.loadingText}>Processing</Typography>
          </Grid>
        </Grid>
        <VSpace units={15}/>
      </>
    );
  }

  const initialValues = {
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
                email: me?.personalUser?.email, // Stripe receipts will be sent to the user's Beneath email address
              }
            }
          });

        if (response.error) {
          actions.setStatus(response.error.message);
        }

        if (response.setupIntent) {
          // sleep so that the Beneath backend can create the billing method and update billing info
          // (another option is to poll until we get back a list of billing methods that is bigger than the original)
          setLoading(true);
          await new Promise(r => setTimeout(r, 3500));
          setLoading(false);

          // updates the billing method list
          client.cache.evict({id: "ROOT_QUERY", fieldName: "billingMethods"});

          // updates the "active" flag
          client.cache.evict({id: "ROOT_QUERY", fieldName: "billingInfo"});

          // close the dialog
          openDialogFn(false);

          // TODO: add an alert at the top of the page "billing method successfully added"
        }

        // NOTES:
        // - run onCompleted() (which should close the dialog and show a success message)

        return;
      }}
    >
      {({ isSubmitting, status }) => (
        <Form title="Add a card">
          <CardElement
            className={classes.cardInput}
            options={{
              style: {
                base: { fontSize: "18px", color: "#FFFFFF" }
              }
            }} />
          <SubmitControl label="Submit" errorAlert={status} disabled={isSubmitting} cancelFn={() => openDialogFn(false)} cancelLabel="Back"/>
        </Form>
      )}
    </Formik>
  );
};

const stripePromise = loadStripe(STRIPE_KEY);

const CardForm: FC<Props> = ({ organization, openDialogFn }) => {
  return (
    <Elements stripe={stripePromise}>
      <CardFormElement organization={organization} openDialogFn={openDialogFn} />
    </Elements>
  );
};

export default CardForm;
