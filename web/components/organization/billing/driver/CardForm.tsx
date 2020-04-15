import React, { FC, useEffect } from 'react'
import { StripeProvider, Elements, injectStripe, CardElement, ReactStripeElements } from 'react-stripe-elements'
import { TextField, Typography, Button, Dialog, DialogActions, DialogContent, Grid } from "@material-ui/core"
import { Autocomplete } from "@material-ui/lab"
import { makeStyles } from "@material-ui/core/styles"
import _ from 'lodash'

import { useToken } from '../../../../hooks/useToken'
import useMe from "../../../../hooks/useMe";
import connection from "../../../../lib/connection"
import billing from "../../../../lib/billing"
import Loading from "../../../Loading"

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

interface CheckoutStateTypes {
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
}

const CardFormWrappedFxn: FC<Props> = ({ stripe }) => {
  const [values, setValues] = React.useState<CheckoutStateTypes>({
    city: "",
    country: "",
    line1: "",
    line2: "",
    postalCode: "",
    state: "",
    email: "",
    cardholder: "",
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
              onClick={() => {
                // refresh page, which should bring user back to either the BillingPlanMenu or the CardDetails
                window.location.reload(true)
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

// The following uses the injectStripe higher-order component to inject the "stripe" object. This is necessary when doing server-side rendering.
// see Stripe React docs: https://github.com/stripe/react-stripe-elements/blob/master/README.md
// see React HOC docs: https://reactjs.org/docs/higher-order-components.html

// * convert the CardFormWrappedFxn functional component to the CardFormWrappedCls class component, so that we can use the injectStripe HOC * //
// HOCs are only for class components

class CardFormWrappedCls extends React.Component<ReactStripeElements.InjectedStripeProps> {
  constructor(props: ReactStripeElements.InjectedStripeProps) {
    super(props);
  }

  render() {
    return <CardFormWrappedFxn stripe={this.props.stripe} />
  }
}

// * apply the injectStripe() HOC * //
// the HOC (injectStripe) is a function that takes one component as an input (CardFormWrappedCls) and returns a new component (CardFormInjectedStripe)
// injectStripe is a function that passes the wrapped component the "stripe" object
const CardFormInjectedStripe = injectStripe(CardFormWrappedCls)

// * set our Stripe key and return the CardForm * //
interface CardFormProps {
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
          <CardFormInjectedStripe />
        </Elements>
      </StripeProvider>
    );
  }
}

export default CardForm;