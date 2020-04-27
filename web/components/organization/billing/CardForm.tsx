import React, { FC, useEffect } from 'react'
import { StripeProvider, Elements, injectStripe, CardElement, ReactStripeElements } from 'react-stripe-elements'
import { TextField, Typography, Button, Dialog, DialogActions, DialogContent, Grid } from "@material-ui/core"
import { Autocomplete } from "@material-ui/lab"
import { makeStyles } from "@material-ui/core/styles"
import _ from 'lodash'

import { useToken } from '../../../hooks/useToken'
import useMe from "../../../hooks/useMe";
import connection from "../../../lib/connection"
import billing from "../../../lib/billing"
import Loading from "../../Loading"

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
    '& > span': {
      marginRight: 10,
      fontSize: 18,
    },
    color: "white"
  },
}))

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
    }
  }
}

interface CardFormStateTypes {
  city: string,
  country: string,
  line1: string,
  line2: string,
  postalCode: string,
  state: string,
  email: string,
  cardholder: string,
  formSubmit: number,
  dialog: boolean,
  error: string | undefined,
  stripeError: string | undefined,
  loading: boolean,
  intentLoading: boolean,
  status: stripe.setupIntents.SetupIntentStatus | null,
}

interface Props {
  stripe: ReactStripeElements.StripeProps | undefined
  closeDialogue: () => void
}

const CardFormWrappedFxn: FC<Props> = ({ stripe, closeDialogue }) => {
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
  })
  const token = useToken()
  const classes = useStyles()

  // get me for email address and organizationID
  const me = useMe();
  if (!me) {
    return <p>Need to log in to proceed to payment</p>
  }

  // When card form is submitted, initiate setupIntent
  useEffect(() => {
    let isMounted = true

    const fetchData = (async () => {
      if (!stripe) {
        return
      }
      if (!values.country) {
        setValues({ ...values, ...{ stripeError: "Missing country", intentLoading: false } })
        return
      }

      const headers = { authorization: `Bearer ${token}` }
      let url = `${connection.API_URL}/billing/stripecard/generate_setup_intent`
      url += `?organizationID=${me.billingOrganization.organizationID}`
      const res = await fetch(url, { headers })

      if (isMounted && me) {
        const intent: any = await res.json()

        if (!res.ok) {
          setValues({ ...values, ...{ error: intent.error, intentLoading: false } })
          return
        }

        const customerData: PaymentMethodData = {
          payment_method_data: {
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
        }

        // handleCardSetup automatically pulls credit card info from the Card element
        // TODO from Stripe Docs: stripe.handleCardSetup may trigger a 3D Secure authentication challenge.This will be shown in a modal dialog and may be confusing for customers using assistive technologies like screen readers.You should make your form accessible by ensuring that success or error messages are clearly read out after this method completes
        const result: stripe.SetupIntentResponse = await stripe.handleCardSetup(intent.client_secret, customerData)
        if (result.error) {
          setValues({ ...values, ...{ stripeError: result.error.message, intentLoading: false } })
        }
        if (result.setupIntent) {
          setValues({ ...values, ...{ status: result.setupIntent.status, intentLoading: false } })
        }
      }
    })

    fetchData()

    // avoid memory leak when component unmounts
    return () => {
      isMounted = false
    }
  }, [values.formSubmit]) // Q: check to see if this useEffect is getting triggered on load

  // Handle submission of Card Details Form
  const handleChange = (name: string) => (event: any) => {
    setValues({ ...values, [name]: event.target.value })
  }

  const onCountryChange = (object: any, value: any) => {
    if (value) {
      setValues({ ...values, country: value.code })
    }
  }

  const handleFormSubmit = (ev: any) => {
    // We don't want to let default form submission happen here, which would refresh the page.
    ev.preventDefault()
    setValues({ ...values, ...{ formSubmit: values.formSubmit + 1, stripeError: "", dialog: true, intentLoading: true } })
    return
  }

  const handleDialogClose = async () => {
    if (values.stripeError || values.error) {
      setValues({ ...values, ...{ dialog: false } })
    }
    if (values.status !== null && values.status === "succeeded") {
      // wait a second so that we can process Stripe's response and show the user their new billing plan
      await new Promise(r => setTimeout(r, 1000));
      
      // reload the page to get new customer billing info from Stripe
      window.location.reload(true)
    }
  }

  if (values.loading) {
    return <Loading justify="center" />
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
              getOptionLabel={option => option.label}
              renderOption={option => (
                <React.Fragment>
                  {option.label}
                </React.Fragment>
              )}
              onChange={onCountryChange}
              renderInput={params => (
                <TextField
                  {...params}
                  label="Choose a country"
                  variant="outlined"
                  fullWidth
                  inputProps={{
                    ...params.inputProps,
                    autoComplete: 'new-password', // disable autocomplete and autofill
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
            <CardElement style={{ base: { fontSize: '18px', color: '#FFFFFF' } }} />
          </Grid>
        </Grid>
        <Grid container className={classes.buttons} spacing={2}>
          <Grid item>
            <Button
              variant="contained"
              onClick={() => { closeDialogue() }}>
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
  )
}

// The following uses the injectStripe higher-order component to inject the "stripe" object. This is necessary when doing server-side rendering.
// see Stripe React docs: https://github.com/stripe/react-stripe-elements/blob/master/README.md
// see React HOC docs: https://reactjs.org/docs/higher-order-components.html

// * convert the CardFormWrappedFxn functional component to the CardFormWrappedCls class component, so that we can use the injectStripe HOC * //
// HOCs are only for class components
class CardFormWrappedCls extends React.Component<ReactStripeElements.InjectedStripeProps & CardFormProps> {
  constructor(props: ReactStripeElements.InjectedStripeProps & CardFormProps) {
    super(props);
  }

  render() {
    return <CardFormWrappedFxn stripe={this.props.stripe} closeDialogue={this.props.closeDialogue} />
  }
}

// * apply the injectStripe() HOC * //
// the HOC (injectStripe) is a function that takes one component as an input (CardFormWrappedCls) and returns a new component (CardFormInjectedStripe)
// injectStripe is a function that passes the wrapped component the "stripe" object
const CardFormInjectedStripe = injectStripe(CardFormWrappedCls)

// * set our Stripe key and return the CardForm * //
interface CardFormProps {
  closeDialogue: () => void
}

interface CardFormState {
  stripe: stripe.Stripe | null;
}

class CardForm extends React.Component<CardFormProps, CardFormState> {
  constructor(props: CardFormProps) {
    super(props);
    this.state = { stripe: null };
  }

  componentDidMount() {
    // Create Stripe instance in componentDidMount (componentDidMount only fires in browser/DOM environment) 
    // note that updating the state like this will cause the CardForm to fire/initially render twice
    this.setState({ stripe: window.Stripe(billing.STRIPE_KEY) })
  }

  render() {
    return (
      <StripeProvider stripe={this.state.stripe}>
        <Elements>
          <CardFormInjectedStripe closeDialogue={this.props.closeDialogue} />
        </Elements>
      </StripeProvider>
    );
  }
}

export default CardForm;