import React, { FC, useEffect } from 'react'
import { CardElement } from 'react-stripe-elements'
import Loading from "../Loading"
import { TextField, Typography, Button } from "@material-ui/core"
import { Autocomplete } from "@material-ui/lab"
import { makeStyles } from "@material-ui/core/styles"
import Grid from '@material-ui/core/Grid'
import { useToken } from '../../hooks/useToken'
import { ReactStripeElements } from 'react-stripe-elements'
import connection from "../../lib/connection"
import _ from 'lodash'
import { useQuery } from "@apollo/react-hooks"
import { QUERY_ME } from "../../apollo/queries/user"
import { Me } from "../../apollo/types/Me"
import Dialog from '@material-ui/core/Dialog'
import DialogActions from '@material-ui/core/DialogActions'
import DialogContent from '@material-ui/core/DialogContent'
import DialogContentText from '@material-ui/core/DialogContentText'
import DialogTitle from '@material-ui/core/DialogTitle'


const PRO_BILLING_PLAN_ID = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"

const useStyles = makeStyles((theme) => ({
  noDataCaption: {
    color: theme.palette.text.secondary,
  },
  title: {
    marginTop: theme.spacing(2),
    marginBottom: theme.spacing(1),
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
  stripeError: stripe.Error | undefined,
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
    stripeError: undefined,
    paymentDetails: null,
    stripeDialog: false,
    loading: false,
    intentLoading: false,
    status: null
  })
  const token = useToken()
  const classes = useStyles()

  // get me for email address
  const { loading, error, data } = useQuery<Me>(QUERY_ME)

  if (loading) {
    return <Loading justify="center" />
  }

  if (error || !data || !data.me) {
    return <p>Error: {JSON.stringify(error)}</p>
  }

  const { me } = data

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
    setValues({ ...values, ...{ cardDetailsFormSubmit: values.cardDetailsFormSubmit + 1, stripeError: undefined, stripeDialog: true, intentLoading: true } })
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
          {values.stripeError && (<Typography variant="body1" color="error">
            {JSON.stringify(values.stripeError.message).substring(1, JSON.stringify(values.stripeError.message).length - 1)}
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
      
      const res = await fetch(url, { headers })

      if (isMounted) {
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
          setValues({ ...values, ...{ stripeError: result.error, intentLoading: false } })
        }
        if (result.setupIntent) {
          console.log(result.setupIntent)
          setValues({ ...values, ...{ stripeError: result.error, status: result.setupIntent.status, intentLoading: false } })
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

  if (values.paymentDetails == null) {
    return <p>Error: no payment details.. why?</p>
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
          <Typography variant="h6" gutterBottom className={classes.title}>
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
        </Grid>
        <Grid item container direction="column" xs={12} sm={6}>
          <Typography variant="h6" gutterBottom className={classes.title}>
            Payment details
            </Typography>
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
      <Grid container>
        <Grid container item xs={12} sm={6} spacing={2}>
          <Grid item>
          <Typography variant="body1" >Interested in upgrading to an Enterprise plan?</Typography>
          </Grid>
          <Grid item>
          <Button
            variant="contained"
            color="primary"
            onClick={() => {
              // TODO: fetch about.beneath.com/contact/demo
            }}>
            Contact Us
          </Button>
          </Grid>
        </Grid>
        <Grid item xs={12} sm={6}>
      <Button
          color="secondary"
          onClick={() => {
            setValues({ ...values, ...{ isSubmittingInfo: true } })
          }}>
          Change card on file
        </Button>
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

// const validateCountry = (val: string) => {
//   return val && val.length == 2
// }

const countries = [
  { code: 'AD', label: 'Andorra'},
  { code: 'AE', label: 'United Arab Emirates'},
  { code: 'AF', label: 'Afghanistan'},
  { code: 'AG', label: 'Antigua and Barbuda'},
  { code: 'AI', label: 'Anguilla'},
  { code: 'AL', label: 'Albania'},
  { code: 'AM', label: 'Armenia'},
  { code: 'AO', label: 'Angola'},
  { code: 'AQ', label: 'Antarctica'},
  { code: 'AR', label: 'Argentina'},
  { code: 'AS', label: 'American Samoa'},
  { code: 'AT', label: 'Austria'},
  { code: 'AU', label: 'Australia'},
  { code: 'AW', label: 'Aruba'},
  { code: 'AX', label: 'Alland Islands'},
  { code: 'AZ', label: 'Azerbaijan'},
  { code: 'BA', label: 'Bosnia and Herzegovina'},
  { code: 'BB', label: 'Barbados'},
  { code: 'BD', label: 'Bangladesh'},
  { code: 'BE', label: 'Belgium'},
  { code: 'BF', label: 'Burkina Faso'},
  { code: 'BG', label: 'Bulgaria'},
  { code: 'BH', label: 'Bahrain'},
  { code: 'BI', label: 'Burundi'},
  { code: 'BJ', label: 'Benin'},
  { code: 'BL', label: 'Saint Barthelemy'},
  { code: 'BM', label: 'Bermuda'},
  { code: 'BN', label: 'Brunei Darussalam'},
  { code: 'BO', label: 'Bolivia'},
  { code: 'BR', label: 'Brazil'},
  { code: 'BS', label: 'Bahamas'},
  { code: 'BT', label: 'Bhutan'},
  { code: 'BV', label: 'Bouvet Island'},
  { code: 'BW', label: 'Botswana'},
  { code: 'BY', label: 'Belarus'},
  { code: 'BZ', label: 'Belize'},
  { code: 'CA', label: 'Canada'},
  { code: 'CC', label: 'Cocos (Keeling) Islands'},
  { code: 'CD', label: 'Congo, Republic of the'},
  { code: 'CF', label: 'Central African Republic'},
  { code: 'CG', label: 'Congo, Democratic Republic of the'},
  { code: 'CH', label: 'Switzerland'},
  { code: 'CI', label: "Cote d'Ivoire"},
  { code: 'CK', label: 'Cook Islands'},
  { code: 'CL', label: 'Chile'},
  { code: 'CM', label: 'Cameroon'},
  { code: 'CN', label: 'China'},
  { code: 'CO', label: 'Colombia'},
  { code: 'CR', label: 'Costa Rica'},
  { code: 'CU', label: 'Cuba'},
  { code: 'CV', label: 'Cape Verde'},
  { code: 'CW', label: 'Curacao'},
  { code: 'CX', label: 'Christmas Island'},
  { code: 'CY', label: 'Cyprus'},
  { code: 'CZ', label: 'Czech Republic'},
  { code: 'DE', label: 'Germany'},
  { code: 'DJ', label: 'Djibouti'},
  { code: 'DK', label: 'Denmark'},
  { code: 'DM', label: 'Dominica'},
  { code: 'DO', label: 'Dominican Republic'},
  { code: 'DZ', label: 'Algeria'},
  { code: 'EC', label: 'Ecuador'},
  { code: 'EE', label: 'Estonia'},
  { code: 'EG', label: 'Egypt'},
  { code: 'EH', label: 'Western Sahara'},
  { code: 'ER', label: 'Eritrea'},
  { code: 'ES', label: 'Spain'},
  { code: 'ET', label: 'Ethiopia'},
  { code: 'FI', label: 'Finland'},
  { code: 'FJ', label: 'Fiji'},
  { code: 'FK', label: 'Falkland Islands (Malvinas)'},
  { code: 'FM', label: 'Micronesia, Federated States of'},
  { code: 'FO', label: 'Faroe Islands'},
  { code: 'FR', label: 'France'},
  { code: 'GA', label: 'Gabon'},
  { code: 'GB', label: 'United Kingdom'},
  { code: 'GD', label: 'Grenada'},
  { code: 'GE', label: 'Georgia'},
  { code: 'GF', label: 'French Guiana'},
  { code: 'GG', label: 'Guernsey'},
  { code: 'GH', label: 'Ghana'},
  { code: 'GI', label: 'Gibraltar'},
  { code: 'GL', label: 'Greenland'},
  { code: 'GM', label: 'Gambia'},
  { code: 'GN', label: 'Guinea'},
  { code: 'GP', label: 'Guadeloupe'},
  { code: 'GQ', label: 'Equatorial Guinea'},
  { code: 'GR', label: 'Greece'},
  { code: 'GS', label: 'South Georgia and the South Sandwich Islands'},
  { code: 'GT', label: 'Guatemala'},
  { code: 'GU', label: 'Guam'},
  { code: 'GW', label: 'Guinea-Bissau'},
  { code: 'GY', label: 'Guyana'},
  { code: 'HK', label: 'Hong Kong'},
  { code: 'HM', label: 'Heard Island and McDonald Islands'},
  { code: 'HN', label: 'Honduras'},
  { code: 'HR', label: 'Croatia'},
  { code: 'HT', label: 'Haiti'},
  { code: 'HU', label: 'Hungary'},
  { code: 'ID', label: 'Indonesia'},
  { code: 'IE', label: 'Ireland'},
  { code: 'IL', label: 'Israel'},
  { code: 'IM', label: 'Isle of Man'},
  { code: 'IN', label: 'India'},
  { code: 'IO', label: 'British Indian Ocean Territory'},
  { code: 'IQ', label: 'Iraq'},
  { code: 'IR', label: 'Iran, Islamic Republic of'},
  { code: 'IS', label: 'Iceland'},
  { code: 'IT', label: 'Italy'},
  { code: 'JE', label: 'Jersey'},
  { code: 'JM', label: 'Jamaica'},
  { code: 'JO', label: 'Jordan'},
  { code: 'JP', label: 'Japan'},
  { code: 'KE', label: 'Kenya'},
  { code: 'KG', label: 'Kyrgyzstan'},
  { code: 'KH', label: 'Cambodia'},
  { code: 'KI', label: 'Kiribati'},
  { code: 'KM', label: 'Comoros'},
  { code: 'KN', label: 'Saint Kitts and Nevis'},
  { code: 'KP', label: "Korea, Democratic People's Republic of"},
  { code: 'KR', label: 'Korea, Republic of'},
  { code: 'KW', label: 'Kuwait'},
  { code: 'KY', label: 'Cayman Islands'},
  { code: 'KZ', label: 'Kazakhstan'},
  { code: 'LA', label: "Lao People's Democratic Republic"},
  { code: 'LB', label: 'Lebanon'},
  { code: 'LC', label: 'Saint Lucia'},
  { code: 'LI', label: 'Liechtenstein'},
  { code: 'LK', label: 'Sri Lanka'},
  { code: 'LR', label: 'Liberia'},
  { code: 'LS', label: 'Lesotho'},
  { code: 'LT', label: 'Lithuania'},
  { code: 'LU', label: 'Luxembourg'},
  { code: 'LV', label: 'Latvia'},
  { code: 'LY', label: 'Libya'},
  { code: 'MA', label: 'Morocco'},
  { code: 'MC', label: 'Monaco'},
  { code: 'MD', label: 'Moldova, Republic of'},
  { code: 'ME', label: 'Montenegro'},
  { code: 'MF', label: 'Saint Martin (French part)'},
  { code: 'MG', label: 'Madagascar'},
  { code: 'MH', label: 'Marshall Islands'},
  { code: 'MK', label: 'Macedonia, the Former Yugoslav Republic of'},
  { code: 'ML', label: 'Mali'},
  { code: 'MM', label: 'Myanmar'},
  { code: 'MN', label: 'Mongolia'},
  { code: 'MO', label: 'Macao'},
  { code: 'MP', label: 'Northern Mariana Islands'},
  { code: 'MQ', label: 'Martinique'},
  { code: 'MR', label: 'Mauritania'},
  { code: 'MS', label: 'Montserrat'},
  { code: 'MT', label: 'Malta'},
  { code: 'MU', label: 'Mauritius'},
  { code: 'MV', label: 'Maldives'},
  { code: 'MW', label: 'Malawi'},
  { code: 'MX', label: 'Mexico'},
  { code: 'MY', label: 'Malaysia'},
  { code: 'MZ', label: 'Mozambique'},
  { code: 'NA', label: 'Namibia'},
  { code: 'NC', label: 'New Caledonia'},
  { code: 'NE', label: 'Niger'},
  { code: 'NF', label: 'Norfolk Island'},
  { code: 'NG', label: 'Nigeria'},
  { code: 'NI', label: 'Nicaragua'},
  { code: 'NL', label: 'Netherlands'},
  { code: 'NO', label: 'Norway'},
  { code: 'NP', label: 'Nepal'},
  { code: 'NR', label: 'Nauru'},
  { code: 'NU', label: 'Niue'},
  { code: 'NZ', label: 'New Zealand'},
  { code: 'OM', label: 'Oman'},
  { code: 'PA', label: 'Panama'},
  { code: 'PE', label: 'Peru'},
  { code: 'PF', label: 'French Polynesia'},
  { code: 'PG', label: 'Papua New Guinea'},
  { code: 'PH', label: 'Philippines'},
  { code: 'PK', label: 'Pakistan'},
  { code: 'PL', label: 'Poland'},
  { code: 'PM', label: 'Saint Pierre and Miquelon'},
  { code: 'PN', label: 'Pitcairn'},
  { code: 'PR', label: 'Puerto Rico'},
  { code: 'PS', label: 'Palestine, State of'},
  { code: 'PT', label: 'Portugal'},
  { code: 'PW', label: 'Palau'},
  { code: 'PY', label: 'Paraguay'},
  { code: 'QA', label: 'Qatar'},
  { code: 'RE', label: 'Reunion'},
  { code: 'RO', label: 'Romania'},
  { code: 'RS', label: 'Serbia'},
  { code: 'RU', label: 'Russian Federation'},
  { code: 'RW', label: 'Rwanda'},
  { code: 'SA', label: 'Saudi Arabia'},
  { code: 'SB', label: 'Solomon Islands'},
  { code: 'SC', label: 'Seychelles'},
  { code: 'SD', label: 'Sudan'},
  { code: 'SE', label: 'Sweden'},
  { code: 'SG', label: 'Singapore'},
  { code: 'SH', label: 'Saint Helena'},
  { code: 'SI', label: 'Slovenia'},
  { code: 'SJ', label: 'Svalbard and Jan Mayen'},
  { code: 'SK', label: 'Slovakia'},
  { code: 'SL', label: 'Sierra Leone'},
  { code: 'SM', label: 'San Marino'},
  { code: 'SN', label: 'Senegal'},
  { code: 'SO', label: 'Somalia'},
  { code: 'SR', label: 'Suriname'},
  { code: 'SS', label: 'South Sudan'},
  { code: 'ST', label: 'Sao Tome and Principe'},
  { code: 'SV', label: 'El Salvador'},
  { code: 'SX', label: 'Sint Maarten (Dutch part)'},
  { code: 'SY', label: 'Syrian Arab Republic'},
  { code: 'SZ', label: 'Swaziland'},
  { code: 'TC', label: 'Turks and Caicos Islands'},
  { code: 'TD', label: 'Chad'},
  { code: 'TF', label: 'French Southern Territories'},
  { code: 'TG', label: 'Togo'},
  { code: 'TH', label: 'Thailand'},
  { code: 'TJ', label: 'Tajikistan'},
  { code: 'TK', label: 'Tokelau'},
  { code: 'TL', label: 'Timor-Leste'},
  { code: 'TM', label: 'Turkmenistan'},
  { code: 'TN', label: 'Tunisia'},
  { code: 'TO', label: 'Tonga'},
  { code: 'TR', label: 'Turkey'},
  { code: 'TT', label: 'Trinidad and Tobago'},
  { code: 'TV', label: 'Tuvalu'},
  { code: 'TW', label: 'Taiwan, Province of China'},
  { code: 'TZ', label: 'United Republic of Tanzania'},
  { code: 'UA', label: 'Ukraine'},
  { code: 'UG', label: 'Uganda'},
  { code: 'US', label: 'United States'},
  { code: 'UY', label: 'Uruguay'},
  { code: 'UZ', label: 'Uzbekistan'},
  { code: 'VA', label: 'Holy See (Vatican City State)'},
  { code: 'VC', label: 'Saint Vincent and the Grenadines'},
  { code: 'VE', label: 'Venezuela'},
  { code: 'VG', label: 'British Virgin Islands'},
  { code: 'VI', label: 'US Virgin Islands'},
  { code: 'VN', label: 'Vietnam'},
  { code: 'VU', label: 'Vanuatu'},
  { code: 'WF', label: 'Wallis and Futuna'},
  { code: 'WS', label: 'Samoa'},
  { code: 'XK', label: 'Kosovo'},
  { code: 'YE', label: 'Yemen'},
  { code: 'YT', label: 'Mayotte'},
  { code: 'ZA', label: 'South Africa'},
  { code: 'ZM', label: 'Zambia'},
  { code: 'ZW', label: 'Zimbabwe'},
]

export default PaymentsByCard