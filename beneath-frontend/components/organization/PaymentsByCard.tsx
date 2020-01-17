import React, { FC, useEffect } from 'react'
import { CardElement } from 'react-stripe-elements'
import Loading from "../Loading"
import { TextField, Typography, Button } from "@material-ui/core"
import { Autocomplete } from "@material-ui/lab"
import { makeStyles } from "@material-ui/core/styles"
import Grid from '@material-ui/core/Grid'
import { useToken } from '../../hooks/useToken'
import useMe from "../../hooks/useMe";
import { ReactStripeElements } from 'react-stripe-elements'
import connection from "../../lib/connection"
import _ from 'lodash'
import Dialog from '@material-ui/core/Dialog'
import DialogActions from '@material-ui/core/DialogActions'
import DialogContent from '@material-ui/core/DialogContent'

const PRO_BILLING_PLAN_ID = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"

const useStyles = makeStyles((theme) => ({
  noDataCaption: {
    color: theme.palette.text.secondary,
  },
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

interface CardPaymentDetails {
  data: {
    organization_id: string,
    card: {
      Brand: string,
      Last4: string,
      ExpMonth: number,
      ExpYear: number,
    },
    billing_details: {
      Name: string,
      Email: string,
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
    }
  }
}

interface CheckoutStateTypes {
  isSubmittingInfo: boolean,
  city: string,
  country: string,
  line1: string,
  line2: string,
  postal_code: string,
  state: string,
  email: string,
  cardholder: string,
  cardDetailsFormSubmit: number,
  stripeError: string | undefined,
  paymentDetails: CardPaymentDetails | null,
  stripeDialog: boolean,
  loading: boolean,
  intentLoading: boolean,
  status: stripe.setupIntents.SetupIntentStatus | null,
}

interface Props {
  stripe: ReactStripeElements.StripeProps | undefined
  organization_id: any
  billing_period: any
  description: any
}

const PaymentsByCard: FC<Props> = ({ stripe, organization_id, billing_period, description }) => {
  const [values, setValues] = React.useState<CheckoutStateTypes>({
    isSubmittingInfo: false,
    city: "",
    country: "",
    line1: "",
    line2: "",
    postal_code: "",
    state: "",
    email: "",
    cardholder: "",
    cardDetailsFormSubmit: 0,
    stripeError: "",
    paymentDetails: null,
    stripeDialog: false,
    loading: false,
    intentLoading: false,
    status: null
  })
  const token = useToken()
  const classes = useStyles()

  // get me for email address
  const me = useMe();

  // Handle submission of Card Details Form
  const handleChange = (name: string) => (event: any) => {
    setValues({ ...values, [name]: event.target.value })
  }

  const onCountryChange = (object: any, value: any) => {
    if (value) {
      setValues({ ...values, country: value.code })
    }
  }

  const handleCardDetailsFormSubmit = (ev: any) => {
    // We don't want to let default form submission happen here, which would refresh the page.
    ev.preventDefault()
    setValues({ ...values, ...{ cardDetailsFormSubmit: values.cardDetailsFormSubmit + 1, stripeError: "", stripeDialog: true, intentLoading: true } })
    return
  }

  const handleDialogClose = () => {
    if (values.stripeError) {
      setValues({ ...values, ...{ stripeDialog: false } })
    }
    if (values.status !== null && values.status === "succeeded") {
      // possibly wait a few seconds
      // reload the page to get new customer billing info from Stripe
      window.location.reload(true)
    }
  }

  const CardBillingDetailsForm = (
    <React.Fragment>
      <form onSubmit={handleCardDetailsFormSubmit}>
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
              value={values.postal_code}
              onChange={handleChange("postal_code")} 
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
              options={countries}
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
            Payment method
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
              onClick={() => {
                setValues({ ...values, ...{ isSubmittingInfo: false } })
              }}>
              Back
            </Button>
          </Grid>
          <Grid item>
            <Button variant="contained" type="submit" color="primary">Submit</Button>
          </Grid>
        </Grid>
      </form>
      <Dialog
        open={values.stripeDialog}
        onClose={handleDialogClose}
        aria-describedby="alert-dialog-description"
      >
        <DialogContent>
          {values.intentLoading && (<Loading />)}
          {values.stripeError && (
            <Typography variant="body1" color="error">
              {values.stripeError}
            </Typography>)}
          {values.status !== null && values.status === "succeeded" && (
            <React.Fragment>
              <Typography variant="h5" gutterBottom>
                Thank you for your order.
              </Typography>
              <Typography variant="subtitle1">
                We will send your bill to your email on file at the beginning of each billing cycle.
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

  // When card form is submitted, initiate setupIntent
  const headers = { authorization: `Bearer ${token}` }
  let url = `${connection.API_URL}/billing/stripecard/generate_setup_intent`
  url += `?organizationID=${organization_id}`
  url += `&billingPlanID=${PRO_BILLING_PLAN_ID}`

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
      
      const res = await fetch(url, { headers })

      if (isMounted && me) {
        if (!res.ok) {
          setValues({ ...values, ...{ error: res.statusText } })
        }
        const intent: any = await res.json()

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
  }, [values.cardDetailsFormSubmit])

  // for paying customers, get current payment details 
  useEffect(() => {
    let isMounted = true
    
    const fetchData = async () => {
      setValues({ ...values, ...{ loading: true } })
      let payment_details_url = `${connection.API_URL}/billing/stripecard/get_payment_details`
      const res = await fetch(payment_details_url, { headers })

      if (isMounted) {
        if (!res.ok) {
          setValues({ ...values, ...{ error: res.statusText } })
        }

        const paymentDetails: CardPaymentDetails = await res.json()
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

  // TODO: this is getting hit "before flicker"
  if (values.paymentDetails == null) {
    return <p></p>
  }

  const address = [values.paymentDetails.data.billing_details.Address.Line1,
    values.paymentDetails.data.billing_details.Address.Line2,
    values.paymentDetails.data.billing_details.Address.City,
    values.paymentDetails.data.billing_details.Address.State,
    values.paymentDetails.data.billing_details.Address.PostalCode,
    values.paymentDetails.data.billing_details.Address.Country].filter(Boolean) // omit Line2 if it's empty
  const payments = [
    { name: 'Card type', detail: _.startCase(_.toLower(values.paymentDetails.data.card.Brand))},
    { name: 'Card number', detail: 'xxxx-xxxx-xxxx-' + values.paymentDetails.data.card.Last4 },
    { name: 'Expiration', detail: values.paymentDetails.data.card.ExpMonth.toString() + '/' + values.paymentDetails.data.card.ExpYear.toString().substring(2,4) },
    { name: 'Card holder', detail: values.paymentDetails.data.billing_details.Name },
    { name: 'Billing address', detail: address.join(', ')}
  ]
  const planDetails = [
    { name: 'Plan name', detail: description },
    { name: 'Billing cycle', detail: billing_period},
  ]

  // current card details
  const CardBillingDetails = (
    <React.Fragment>
      <Grid container spacing={2}>
        <Grid item container direction="column" xs={12} sm={6}>
          <Typography variant="h6" className={classes.title}>
            Billing plan
            </Typography>
          <Grid container>
            {planDetails.map(planDetails => (
              <React.Fragment key={planDetails.name}>
                <Grid item xs={6}>
                  <Typography gutterBottom>{planDetails.name}</Typography>
                </Grid>
                <Grid item xs={6}>
                  <Typography gutterBottom>{planDetails.detail}</Typography>
                </Grid>
              </React.Fragment>
            ))}
          </Grid>
          {/* <Grid item>
            <Button
              color="primary"
              onClick={() => {
                // TODO: fetch about.beneath.com/contact/demo
              }}>
              Contact Us To Upgrade to An Enterprise Plan
            </Button>
          </Grid> */}
        </Grid>
        <Grid item container direction="column" xs={12} sm={6}>
          <Grid container alignItems="center" justify="space-between">
            <Grid item>
              <Typography variant="h6" className={classes.title}>
                Payment details
              </Typography>
            </Grid>
            <Grid item>
              <Button
                color="primary"
                onClick={() => {
                  setValues({ ...values, ...{ isSubmittingInfo: true } })
                }}>
                Edit
              </Button>
            </Grid>
          </Grid>
          <Grid container>
            {payments.map(payments => (
              <React.Fragment key={payments.name}>
                <Grid item xs={6}>
                  <Typography gutterBottom>{payments.name}</Typography>
                </Grid>
                <Grid item xs={6}>
                  <Typography gutterBottom>{payments.detail}</Typography>
                </Grid>
              </React.Fragment>
            ))}
          </Grid>
        </Grid>
      </Grid>
    </React.Fragment>
  )

  // either show input form or show current billing details
  if (values.isSubmittingInfo === true) {
    return CardBillingDetailsForm
  } else {
    return CardBillingDetails
  }
}

const countries = [
  { code: 'AF', label: 'Afghanistan' },
  { code: 'AX', label: 'Åland Islands' },
  { code: 'AL', label: 'Albania' },
  { code: 'DZ', label: 'Algeria' },
  { code: 'AS', label: 'American Samoa' },
  { code: 'AD', label: 'Andorra' },
  { code: 'AO', label: 'Angola' },
  { code: 'AI', label: 'Anguilla' },
  { code: 'AQ', label: 'Antarctica' },
  { code: 'AG', label: 'Antigua and Barbuda' },
  { code: 'AR', label: 'Argentina' },
  { code: 'AM', label: 'Armenia' },
  { code: 'AW', label: 'Aruba' },
  { code: 'AU', label: 'Australia' },
  { code: 'AT', label: 'Austria' },
  { code: 'AZ', label: 'Azerbaijan' },
  { code: 'BS', label: 'Bahamas' },
  { code: 'BH', label: 'Bahrain' },
  { code: 'BD', label: 'Bangladesh' },
  { code: 'BB', label: 'Barbados' },
  { code: 'BY', label: 'Belarus' },
  { code: 'BE', label: 'Belgium' },
  { code: 'BZ', label: 'Belize' },
  { code: 'BJ', label: 'Benin' },
  { code: 'BM', label: 'Bermuda' },
  { code: 'BT', label: 'Bhutan' },
  { code: 'BO', label: 'Bolivia (Plurinational State of)' },
  { code: 'BQ', label: 'Bonaire, Sint Eustatius and Saba' },
  { code: 'BA', label: 'Bosnia and Herzegovina' },
  { code: 'BW', label: 'Botswana' },
  { code: 'BV', label: 'Bouvet Island' },
  { code: 'BR', label: 'Brazil' },
  { code: 'IO', label: 'British Indian Ocean Territory' },
  { code: 'BN', label: 'Brunei Darussalam' },
  { code: 'BG', label: 'Bulgaria' },
  { code: 'BF', label: 'Burkina Faso' },
  { code: 'BI', label: 'Burundi' },
  { code: 'CV', label: 'Cabo Verde' },
  { code: 'KH', label: 'Cambodia' },
  { code: 'CM', label: 'Cameroon' },
  { code: 'CA', label: 'Canada' },
  { code: 'KY', label: 'Cayman Islands' },
  { code: 'CF', label: 'Central African Republic' },
  { code: 'TD', label: 'Chad' },
  { code: 'CL', label: 'Chile' },
  { code: 'CN', label: 'China' },
  { code: 'CX', label: 'Christmas Island' },
  { code: 'CC', label: 'Cocos (Keeling) Islands' },
  { code: 'CO', label: 'Colombia' },
  { code: 'KM', label: 'Comoros' },
  { code: 'CG', label: 'Congo' },
  { code: 'CD', label: 'Congo, Democratic Republic of the' },
  { code: 'CK', label: 'Cook Islands' },
  { code: 'CR', label: 'Costa Rica' },
  { code: 'CI', label: "Côte d'Ivoire"},
  { code: 'HR', label: 'Croatia' },
  { code: 'CU', label: 'Cuba' },
  { code: 'CW', label: 'Curaçao' },
  { code: 'CY', label: 'Cyprus' },
  { code: 'CZ', label: 'Czechia' },
  { code: 'DK', label: 'Denmark' },
  { code: 'DJ', label: 'Djibouti' },
  { code: 'DM', label: 'Dominica' },
  { code: 'DO', label: 'Dominican Republic' },
  { code: 'EC', label: 'Ecuador' },
  { code: 'EG', label: 'Egypt' },
  { code: 'SV', label: 'El Salvador' },
  { code: 'GQ', label: 'Equatorial Guinea' },
  { code: 'ER', label: 'Eritrea' },
  { code: 'EE', label: 'Estonia' },
  { code: 'SZ', label: 'Eswatini' },
  { code: 'ET', label: 'Ethiopia' },
  { code: 'FK', label: 'Falkland Islands (Malvinas)' },
  { code: 'FO', label: 'Faroe Islands' },
  { code: 'FJ', label: 'Fiji' },
  { code: 'FI', label: 'Finland' },
  { code: 'FR', label: 'France' },
  { code: 'GF', label: 'French Guiana' },
  { code: 'PF', label: 'French Polynesia' },
  { code: 'TF', label: 'French Southern Territories' },
  { code: 'GA', label: 'Gabon' },
  { code: 'GM', label: 'Gambia' },
  { code: 'GE', label: 'Georgia' },
  { code: 'DE', label: 'Germany' },
  { code: 'GH', label: 'Ghana' },
  { code: 'GI', label: 'Gibraltar' },
  { code: 'GR', label: 'Greece' },
  { code: 'GL', label: 'Greenland' },
  { code: 'GD', label: 'Grenada' },
  { code: 'GP', label: 'Guadeloupe' },
  { code: 'GU', label: 'Guam' },
  { code: 'GT', label: 'Guatemala' },
  { code: 'GG', label: 'Guernsey' },
  { code: 'GN', label: 'Guinea' },
  { code: 'GW', label: 'Guinea-Bissau' },
  { code: 'GY', label: 'Guyana' },
  { code: 'HT', label: 'Haiti' },
  { code: 'HM', label: 'Heard Island and McDonald Islands' },
  { code: 'VA', label: 'Holy See' },
  { code: 'HN', label: 'Honduras' },
  { code: 'HK', label: 'Hong Kong' },
  { code: 'HU', label: 'Hungary' },
  { code: 'IS', label: 'Iceland' },
  { code: 'IN', label: 'India' },
  { code: 'ID', label: 'Indonesia' },
  { code: 'IR', label: 'Iran (Islamic Republic of)' },
  { code: 'IQ', label: 'Iraq' },
  { code: 'IE', label: 'Ireland' },
  { code: 'IM', label: 'Isle of Man' },
  { code: 'IL', label: 'Israel' },
  { code: 'IT', label: 'Italy' },
  { code: 'JM', label: 'Jamaica' },
  { code: 'JP', label: 'Japan' },
  { code: 'JE', label: 'Jersey' },
  { code: 'JO', label: 'Jordan' },
  { code: 'KZ', label: 'Kazakhstan' },
  { code: 'KE', label: 'Kenya' },
  { code: 'KI', label: 'Kiribati' },
  { code: 'KP', label: "Korea (Democratic People's Republic of) "},
  { code: 'KR', label: 'Korea, Republic of' },
  { code: 'KW', label: 'Kuwait' },
  { code: 'KG', label: 'Kyrgyzstan' },
  { code: 'LA', label: "Lao People's Democratic Republic"},
  { code: 'LV', label: 'Latvia' },
  { code: 'LB', label: 'Lebanon' },
  { code: 'LS', label: 'Lesotho' },
  { code: 'LR', label: 'Liberia' },
  { code: 'LY', label: 'Libya' },
  { code: 'LI', label: 'Liechtenstein' },
  { code: 'LT', label: 'Lithuania' },
  { code: 'LU', label: 'Luxembourg' },
  { code: 'MO', label: 'Macao' },
  { code: 'MG', label: 'Madagascar' },
  { code: 'MW', label: 'Malawi' },
  { code: 'MY', label: 'Malaysia' },
  { code: 'MV', label: 'Maldives' },
  { code: 'ML', label: 'Mali' },
  { code: 'MT', label: 'Malta' },
  { code: 'MH', label: 'Marshall Islands' },
  { code: 'MQ', label: 'Martinique' },
  { code: 'MR', label: 'Mauritania' },
  { code: 'MU', label: 'Mauritius' },
  { code: 'YT', label: 'Mayotte' },
  { code: 'MX', label: 'Mexico' },
  { code: 'FM', label: 'Micronesia (Federated States of)' },
  { code: 'MD', label: 'Moldova, Republic of' },
  { code: 'MC', label: 'Monaco' },
  { code: 'MN', label: 'Mongolia' },
  { code: 'ME', label: 'Montenegro' },
  { code: 'MS', label: 'Montserrat' },
  { code: 'MA', label: 'Morocco' },
  { code: 'MZ', label: 'Mozambique' },
  { code: 'MM', label: 'Myanmar' },
  { code: 'NA', label: 'Namibia' },
  { code: 'NR', label: 'Nauru' },
  { code: 'NP', label: 'Nepal' },
  { code: 'NL', label: 'Netherlands' },
  { code: 'NC', label: 'New Caledonia' },
  { code: 'NZ', label: 'New Zealand' },
  { code: 'NI', label: 'Nicaragua' },
  { code: 'NE', label: 'Niger' },
  { code: 'NG', label: 'Nigeria' },
  { code: 'NU', label: 'Niue' },
  { code: 'NF', label: 'Norfolk Island' },
  { code: 'MK', label: 'North Macedonia' },
  { code: 'MP', label: 'Northern Mariana Islands' },
  { code: 'NO', label: 'Norway' },
  { code: 'OM', label: 'Oman' },
  { code: 'PK', label: 'Pakistan' },
  { code: 'PW', label: 'Palau' },
  { code: 'PS', label: 'Palestine, State of' },
  { code: 'PA', label: 'Panama' },
  { code: 'PG', label: 'Papua New Guinea' },
  { code: 'PY', label: 'Paraguay' },
  { code: 'PE', label: 'Peru' },
  { code: 'PH', label: 'Philippines' },
  { code: 'PN', label: 'Pitcairn' },
  { code: 'PL', label: 'Poland' },
  { code: 'PT', label: 'Portugal' },
  { code: 'PR', label: 'Puerto Rico' },
  { code: 'QA', label: 'Qatar' },
  { code: 'RE', label: 'Réunion' },
  { code: 'RO', label: 'Romania' },
  { code: 'RU', label: 'Russian Federation' },
  { code: 'RW', label: 'Rwanda' },
  { code: 'BL', label: 'Saint Barthélemy' },
  { code: 'SH', label: 'Saint Helena, Ascension and Tristan da Cunha' },
  { code: 'KN', label: 'Saint Kitts and Nevis' },
  { code: 'LC', label: 'Saint Lucia' },
  { code: 'MF', label: 'Saint Martin (French part)' },
  { code: 'PM', label: 'Saint Pierre and Miquelon' },
  { code: 'VC', label: 'Saint Vincent and the Grenadines' },
  { code: 'WS', label: 'Samoa' },
  { code: 'SM', label: 'San Marino' },
  { code: 'ST', label: 'Sao Tome and Principe' },
  { code: 'SA', label: 'Saudi Arabia' },
  { code: 'SN', label: 'Senegal' },
  { code: 'RS', label: 'Serbia' },
  { code: 'SC', label: 'Seychelles' },
  { code: 'SL', label: 'Sierra Leone' },
  { code: 'SG', label: 'Singapore' },
  { code: 'SX', label: 'Sint Maarten (Dutch part)' },
  { code: 'SK', label: 'Slovakia' },
  { code: 'SI', label: 'Slovenia' },
  { code: 'SB', label: 'Solomon Islands' },
  { code: 'SO', label: 'Somalia' },
  { code: 'ZA', label: 'South Africa' },
  { code: 'GS', label: 'South Georgia and the South Sandwich Islands' },
  { code: 'SS', label: 'South Sudan' },
  { code: 'ES', label: 'Spain' },
  { code: 'LK', label: 'Sri Lanka' },
  { code: 'SD', label: 'Sudan' },
  { code: 'SR', label: 'Suriname' },
  { code: 'SJ', label: 'Svalbard and Jan Mayen' },
  { code: 'SE', label: 'Sweden' },
  { code: 'CH', label: 'Switzerland' },
  { code: 'SY', label: 'Syrian Arab Republic' },
  { code: 'TW', label: 'Taiwan, Province of China' },
  { code: 'TJ', label: 'Tajikistan' },
  { code: 'TZ', label: 'Tanzania, United Republic of' },
  { code: 'TH', label: 'Thailand' },
  { code: 'TL', label: 'Timor-Leste' },
  { code: 'TG', label: 'Togo' },
  { code: 'TK', label: 'Tokelau' },
  { code: 'TO', label: 'Tonga' },
  { code: 'TT', label: 'Trinidad and Tobago' },
  { code: 'TN', label: 'Tunisia' },
  { code: 'TR', label: 'Turkey' },
  { code: 'TM', label: 'Turkmenistan' },
  { code: 'TC', label: 'Turks and Caicos Islands' },
  { code: 'TV', label: 'Tuvalu' },
  { code: 'UG', label: 'Uganda' },
  { code: 'UA', label: 'Ukraine' },
  { code: 'AE', label: 'United Arab Emirates' },
  { code: 'GB', label: 'United Kingdom of Great Britain and Northern Ireland' },
  { code: 'UM', label: 'United States Minor Outlying Islands' },
  { code: 'US', label: 'United States of America' },
  { code: 'UY', label: 'Uruguay' },
  { code: 'UZ', label: 'Uzbekistan' },
  { code: 'VU', label: 'Vanuatu' },
  { code: 'VE', label: 'Venezuela (Bolivarian Republic of)' },
  { code: 'VN', label: 'Viet Nam' },
  { code: 'VG', label: 'Virgin Islands (British)' },
  { code: 'VI', label: 'Virgin Islands (U.S.)' },
  { code: 'WF', label: 'Wallis and Futuna' },
  { code: 'EH', label: 'Western Sahara' },
  { code: 'YE', label: 'Yemen' },
  { code: 'ZM', label: 'Zambia' },
  { code: 'ZW', label: 'Zimbabwe' }
]

export default PaymentsByCard