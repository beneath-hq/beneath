import { Button, Dialog, DialogActions, DialogContent, Grid, TextField, Typography} from "@material-ui/core";
import { makeStyles } from "@material-ui/core/styles";
import { Autocomplete } from "@material-ui/lab";
import _ from "lodash";
import React, { FC, useEffect } from "react";

import { CardElement, Elements, useElements, useStripe } from "@stripe/react-stripe-js";
import { loadStripe } from "@stripe/stripe-js";

import useMe from "../../../hooks/useMe";
import { useToken } from "../../../hooks/useToken";
import billing from "../../../lib/billing";
import connection from "../../../lib/connection";
import Loading from "../../Loading";

const useStyles = makeStyles((theme) => ({
  title: {
    marginTop: theme.spacing(2),
    marginBottom: theme.spacing(2),
  },
  buttons: {
    marginTop: theme.spacing(3),
  },
  input: {
    "&:-webkit-autofill": {
      WebkitBoxShadow: "0 0 0px 1000px rgba(16, 24, 46, 1) inset", // color is from theme.palette.background.default
      WebkitTextFillColor: "white"
    },
    color: theme.palette.text.secondary,
  },
  option: {
    fontSize: 15,
    "& > span": {
      marginRight: 10,
      fontSize: 18,
    },
    color: "white"
  },
}));

interface CardFormStateTypes {
  city: string;
  country: string;
  line1: string;
  line2: string;
  postalCode: string;
  state: string;
  email: string;
  cardholder: string;
  formSubmit: number;
  dialog: boolean;
  error: string | undefined;
  stripeError: string | undefined;
  loading: boolean;
  intentLoading: boolean;
  status: stripe.setupIntents.SetupIntentStatus | null;
}

interface Props {
  closeDialogue: () => void;
}

const CardFormElement: FC<Props> = ({ closeDialogue }) => {
  const [values, setValues] = React.useState<CardFormStateTypes>({
    cardholder: "",
    line1: "",
    line2: "",
    city: "",
    state: "",
    postalCode: "",
    country: "",
    email: "",
    formSubmit: 0,
    dialog: false,
    error: "",
    stripeError: "",
    loading: false,
    intentLoading: false,
    status: null
  });
  const token = useToken();
  const classes = useStyles();
  const me = useMe();
  const stripe = useStripe();
  const elements = useElements();

  // When card form is submitted, initiate setupIntent
  useEffect(() => {
    let isMounted = true;

    const fetchData = (async () => {
      if (!values.country) {
        setValues({ ...values, ...{ stripeError: "Missing country", intentLoading: false } });
        return;
      }

      if (!me) {
        return <p>Need to log in to proceed to payment</p>;
      }

      if (!elements) {
        return <p>Unable to get Stripe elements</p>;
      }

      if (!stripe) {
        return <p>Unable to get Stripe</p>;
      }

      const headers = { authorization: `Bearer ${token}` };
      let url = `${connection.API_URL}/billing/stripecard/generate_setup_intent`;
      url += `?organizationID=${me.billingOrganization.organizationID}`;
      const res = await fetch(url, { headers });

      if (isMounted) {
        const intent: any = await res.json();

        if (!res.ok) {
          setValues({ ...values, ...{ error: intent.error, intentLoading: false } });
          return;
        }

        // handleCardSetup automatically pulls credit card info from the Card element
        // TODO from Stripe Docs: stripe.handleCardSetup may trigger a 3D Secure authentication challenge.
        // This will be shown in a modal dialog and may be confusing for customers using
        // assistive technologies like screen readers. You should make your form accessible by ensuring
        // that success or error messages are clearly read out after this method completes
        const cardElement = elements.getElement(CardElement);
        if (!cardElement) {
          setValues({ ...values, ...{ error: intent.error, intentLoading: false } });
          return;
        }

        const response: stripe.SetupIntentResponse | any =
          await stripe.confirmCardSetup(intent.client_secret, {
            payment_method: {
              card: cardElement,
              billing_details: {
                address: {
                  city: values.city,
                  country: values.country,
                  line1: values.line1,
                  line2: values.line2,
                  postal_code: values.postalCode,
                  state: values.state,
                },
                email: me.email, // Stripe receipts will be sent to the user's Beneath email address
                name: values.cardholder,
              }
            }
          });

        if (response.error) {
          setValues({ ...values, ...{ stripeError: response.error.message, intentLoading: false } });
        }
        if (response.setupIntent) {
          setValues({ ...values, ...{ status: response.setupIntent.status, intentLoading: false } });
        }
      }
    });

    fetchData();

    // avoid memory leak when component unmounts
    return () => {
      isMounted = false;
    };
  }, [values.formSubmit]); // Q: check to see if this useEffect is getting triggered on load

  // Handle submission of Card Details Form
  const handleChange = (name: string) => (event: any) => {
    setValues({ ...values, [name]: event.target.value });
  };

  const onCountryChange = (object: any, value: any) => {
    if (value) {
      setValues({ ...values, country: value.code });
    }
  };

  const handleFormSubmit = (ev: any) => {
    // We don't want to let default form submission happen here, which would refresh the page.
    ev.preventDefault();

    if (!stripe || !elements) {
      // Stripe.js has not loaded yet. Make sure to disable
      // form submission until Stripe.js has loaded.
      return;
    }

    setValues({ ...values,
      ...{ formSubmit: values.formSubmit + 1, stripeError: "", dialog: true, intentLoading: true } });
    return;
  };

  const handleDialogClose = async () => {
    if (values.stripeError || values.error) {
      setValues({ ...values, ...{ dialog: false } });
    }
    if (values.status !== null && values.status === "succeeded") {
      // wait a second so that we can process Stripe's response and show the user their new billing plan
      await new Promise((r) => setTimeout(r, 1000));

      // reload the page to get new customer billing info from Stripe
      window.location.reload(true);
    }
  };

  if (values.loading) {
    return <Loading justify="center" />;
  }

  return (
    <React.Fragment>
      <form onSubmit={handleFormSubmit}>
        <Typography variant="h6" gutterBottom>
          Billing information
        </Typography>
        <Grid container spacing={3}>
          <Grid item xs={12}>
            <TextField
              required
              id="cardholder"
              name="cardholder"
              label="Name on card"
              fullWidth
              autoComplete="billing name"
              inputProps={{
                className: classes.input
              }}
              value={values.cardholder}
              onChange={handleChange("cardholder")}
            />
          </Grid>
          <Grid item xs={12} sm={6}>
            <TextField
              required
              id="address1"
              name="address1"
              label="Address line 1"
              fullWidth
              autoComplete="billing address-line1"
              inputProps={{
                className: classes.input
              }}
              value={values.line1}
              onChange={handleChange("line1")}
            />
          </Grid>
          <Grid item xs={12} sm={6}>
            <TextField
              id="address2"
              name="address2"
              label="Address line 2"
              fullWidth
              autoComplete="billing address-line2"
              inputProps={{
                className: classes.input
              }}
              value={values.line2}
              onChange={handleChange("line2")}
            />
          </Grid>
          <Grid item xs={12} sm={6}>
            <TextField
              required
              id="city"
              name="city"
              label="City"
              fullWidth
              autoComplete="billing address-level2"
              inputProps={{
                className: classes.input
              }}
              value={values.city}
              onChange={handleChange("city")}
            />
          </Grid>
          <Grid item xs={12} sm={6}>
            <TextField
              required
              id="state"
              name="state"
              label="State/Province/Region"
              fullWidth
              autoComplete="billing address-level1"
              inputProps={{
                className: classes.input
              }}
              value={values.state}
              onChange={handleChange("state")}
            />
          </Grid>
          <Grid item xs={12} sm={6}>
            <TextField
              required
              id="zip"
              name="zip"
              label="Zip / Postal code"
              fullWidth
              autoComplete="billing postal-code"
              inputProps={{
                className: classes.input
              }}
              value={values.postalCode}
              onChange={handleChange("postalCode")}
            />
          </Grid>
          {/* <Grid item xs={12} sm={6}>
            <TextField
              required
              id="country"
              name="country"
              label="Country"
              fullWidth
              autoComplete="billing country"
              inputProps={{
                className: classes.input
              }}
              helperText={!validateCountry(values.country) ? "Must be the two letter country code" : undefined}
              value={values.country}
              onChange={handleChange("country")}
            />
          </Grid> */}
          <Grid item xs={12} sm={6}>
            <Autocomplete
              id="country"
              style={{ width: 350 }} // fullWidth // doesn't exist on AutocompleteProps
              options={billing.COUNTRIES}
              classes={{
                option: classes.option,
              }}
              autoHighlight
              getOptionLabel={(option) => option.label}
              renderOption={(option) => (
                <React.Fragment>
                  {option.label}
                </React.Fragment>
              )}
              onChange={onCountryChange}
              renderInput={(params) => (
                <TextField
                  {...params}
                  label="Choose a country"
                  variant="outlined"
                  fullWidth
                  inputProps={{
                    ...params.inputProps,
                    autoComplete: "new-password", // disable autocomplete and autofill
                  }}
                // autoComplete="billing country"
                />
              )}
            />
          </Grid>
        </Grid>
        <Typography variant="h6" gutterBottom className={classes.title}>
          Card details
        </Typography>
        <Grid container>
          <Grid item xs={12} md={6}>
            <CardElement
              options={{
                style: {
                  base: { fontSize: "18px", color: "#FFFFFF" }
                }
              }} />
          </Grid>
        </Grid>
        <Grid container className={classes.buttons} spacing={2}>
          <Grid item>
            <Button
              variant="contained"
              onClick={() => { closeDialogue(); }}>
              Back
            </Button>
          </Grid>
          <Grid item>
            <Button variant="contained" type="submit" color="primary">Submit</Button>
          </Grid>
        </Grid>
      </form>
      <Dialog
        open={values.dialog}
        onClose={handleDialogClose}
        aria-describedby="alert-dialog-description"
      >
        <DialogContent>
          {values.intentLoading && (<Loading />)}
          {values.error && (
            <Typography variant="body1" color="error">
              {values.error}
            </Typography>)}
          {values.stripeError && (
            <Typography variant="body1" color="error">
              {values.stripeError}
            </Typography>)}
          {values.status !== null && values.status === "succeeded" && (
            <React.Fragment>
              <Typography variant="h5" gutterBottom>
                Thank you. Your card has been approved.
              </Typography>
            </React.Fragment>)}
          {values.status !== null && values.status !== "succeeded" && (
            <Typography variant="body1" color="error">
              {values.status}
            </Typography>
          )}
        </DialogContent>
        <DialogActions>
          {!values.intentLoading && (
            <Button onClick={handleDialogClose} color="primary" autoFocus>
              Ok
          </Button>)}
        </DialogActions>
      </Dialog>
    </React.Fragment>
  );
};

const stripePromise = loadStripe(billing.STRIPE_KEY);

const CardForm: FC<Props> = ({ closeDialogue }) => {
  return (
    <Elements stripe={stripePromise}>
      <CardFormElement closeDialogue={closeDialogue} />
    </Elements>
  );
};

export default CardForm;
