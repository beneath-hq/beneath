import React, { FC } from 'react'
import { TextField, Typography, Button, Dialog, DialogActions, DialogContent, Grid, DialogContentText, ListItem, DialogTitle } from "@material-ui/core"
import { Autocomplete } from "@material-ui/lab"
import { makeStyles } from "@material-ui/core/styles"
import _ from 'lodash'

import { useQuery, useMutation } from "@apollo/react-hooks";
import { QUERY_BILLING_INFO, UPDATE_BILLING_INFO } from '../../../apollo/queries/billinginfo';
import { UpdateBillingInfo, UpdateBillingInfoVariables } from '../../../apollo/types/UpdateBillingInfo';
import CheckIcon from '@material-ui/icons/Check';
import SelectField from "../../SelectField";
import billing from "../../../lib/billing"
import { BillingMethods, BillingMethodsVariables } from '../../../apollo/types/BillingMethods'
import { QUERY_BILLING_METHODS } from '../../../apollo/queries/billingmethod'
import { BillingPlans } from '../../../apollo/types/BillingPlans'
import { QUERY_BILLING_PLANS } from '../../../apollo/queries/billingplan'
import { BillingInfo, BillingInfoVariables } from '../../../apollo/types/BillingInfo'

const useStyles = makeStyles((theme) => ({
  firstTitle: {
    marginTop: theme.spacing(0),
    marginBottom: theme.spacing(2),
  },
  title: {
    marginTop: theme.spacing(4),
    marginBottom: theme.spacing(2),
  },
  button: {
    marginTop: theme.spacing(3),
    marginBotton: theme.spacing(2),
    marginRight: theme.spacing(3),
  },
  icon: {
    marginRight: theme.spacing(2),
  },
  selectField: {
    marginTop: theme.spacing(1),
    minWidth: 250,
  },
  textField: {
    marginTop: theme.spacing(2),
    minWidth: 250,
  },
  proratedDescription: {
    marginTop: theme.spacing(3)
  },
  option: {
    fontSize: 15,
    '& > span': {
      marginRight: 10,
      fontSize: 18,
    },
    color: "white"
  },
  input: {
    "&:-webkit-autofill": {
      WebkitBoxShadow: "0 0 0px 1000px rgba(16, 24, 46, 1) inset", // color is from theme.palette.background.default
      WebkitTextFillColor: "white"
    },
    color: theme.palette.text.secondary,
  },
}))

interface CheckoutStateTypes {
  billingMethod: string,
  country: string,
  region: string,
  companyName: string,
  taxID: string,
}

interface Props {
  organizationID: string
  route: string
  closeDialogue: () => void
}

const UpdateBillingInfoDialogue: FC<Props> = ({ organizationID, route, closeDialogue }) => { 
  const classes = useStyles()
  const [successDialogue, setSuccessDialogue] = React.useState(false)
  const [errorDialogue, setErrorDialogue] = React.useState(false)
  const [error, setError] = React.useState("")
  const [values, setValues] = React.useState<CheckoutStateTypes>({
    billingMethod: "",
    country: "",
    region: "",
    companyName: "",
    taxID: "",
  })

  const { loading, error: queryError1, data } = useQuery<BillingInfo, BillingInfoVariables>(QUERY_BILLING_INFO, {
    variables: {
      organizationID: organizationID,
    },
  });

  const { loading: loading2, error: queryError2, data: data2 } = useQuery<BillingMethods, BillingMethodsVariables>(QUERY_BILLING_METHODS, {
    variables: {
      organizationID: organizationID,
    },
  });

  const { loading: loading3, error: queryError3, data: data3 } = useQuery<BillingPlans>(QUERY_BILLING_PLANS);

  const [updateBillingInfo, {error: mutError}] = useMutation<UpdateBillingInfo, UpdateBillingInfoVariables>(UPDATE_BILLING_INFO, {
    onCompleted: (data) => {
      if (data) {
        setSuccessDialogue(true)
      }
    },
    // this doesn't seem to work, but it'd be nice if it did!
    // onError: (error) => {
    //   console.log(error)
    //   setError(error.message.replace("GraphQL error:", ""))
    //   setErrorDialogue(true)
    // },
  })

  if (queryError1 || !data) {
    return <p>Error: {JSON.stringify(queryError1)}</p>;
  }

  if (queryError2 || !data2) {
    return <p>Error: {JSON.stringify(queryError2)}</p>;
  }

  if (queryError3 || !data3) {
    return <p>Error: {JSON.stringify(queryError3)}</p>;
  }

  const cards = data2.billingMethods.filter(billingMethod => billingMethod.paymentsDriver == billing.STRIPECARD_DRIVER)
  const wire = data2.billingMethods.filter(billingMethod => billingMethod.paymentsDriver == billing.STRIPEWIRE_DRIVER)
  const anarchism = data2.billingMethods.filter(billingMethod => billingMethod.paymentsDriver == billing.ANARCHISM_DRIVER)[0]

  const freePlan = data3.billingPlans.filter(billingPlan => billingPlan.default)[0]
  const proPlan = data3.billingPlans.filter(billingPlan => !billingPlan.default)[0]

  const handleChange = (name: string) => (event: any) => {
    setValues({ ...values, [name]: event.target.value })
  }

  const onCountryChange = (object: any, value: any) => {
    if (value) {
      setValues({ ...values, country: value.value })
    }
  }
  
  const onRegionChange = (object: any, value: any) => {
    if (value) {
      setValues({ ...values, region: value.value })
    }
  }

  return (
    <>
      <Dialog
        open={route == "checkout"}
        fullWidth={true}
        maxWidth={"sm"}
        onBackdropClick={() => { closeDialogue() }}
      >
        <DialogTitle id="alert-dialog-title">{"Checkout"}</DialogTitle>
        <DialogContent>
          <Grid container direction="column">
            <Grid item>
              <Typography variant="h2" className={classes.firstTitle}>
                Professional plan: $50/month base
          </Typography>
              <Typography>
                {["5 GB writes included in base. Then $2/GB.", "25 GB reads included in base. Then $1/GB.", "Private projects", "Role-based access controls"].map((feature) => {
                  return (
                    <React.Fragment key={feature}>
                      <ListItem dense>
                        <CheckIcon className={classes.icon} />
                        <Typography component="span">{feature}</Typography>
                      </ListItem>
                    </React.Fragment>
                  )
                })}
              </Typography>
            </Grid>
            <Grid item>
              <Typography variant="h2" className={classes.title}>
                Select your billing method
              </Typography>
              <SelectField
                id="billing_method"
                label="Billing method"
                value={values.billingMethod}
                options={data2.billingMethods.filter(billingMethod => billingMethod.paymentsDriver != billing.ANARCHISM_DRIVER).map((billingMethod) => {
                  if (billingMethod.paymentsDriver == billing.STRIPECARD_DRIVER) {
                    const payload = JSON.parse(billingMethod.driverPayload)
                    return { label: payload.brand.charAt(0).toUpperCase() + payload.brand.slice(1) + " xxxx-xxxx-xxxx-" + payload.last4, value: billingMethod.billingMethodID }
                  } else if (billingMethod.paymentsDriver == billing.STRIPEWIRE_DRIVER) {
                    return { label: "Wire payment", value: billingMethod.billingMethodID }
                  } else {
                    return { label: "", value: "" }
                  }
                })}
                onChange={({ target }) => setValues({ ...values, billingMethod: target.value as string })}
                controlClass={classes.selectField}
              />
            </Grid>
            <Grid item>
              <Typography variant="h2" className={classes.title}>
                Tax information
              </Typography>
              <Grid container spacing={3}>
                <Grid item xs={12} sm={6}>
                  <Autocomplete
                    id="country"
                    style={{ width: 260 }} // fullWidth // doesn't exist on AutocompleteProps
                    options={countries}
                    classes={{option: classes.option,}}
                    className={classes.selectField}
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
                      />
                    )}
                  />
                </Grid>
                {values.country != "United States of America" && (
                  <Grid item xs={12} sm={6}>
                  </Grid>
                )}
                {values.country == "United States of America" && (
                  <Grid item xs={12} sm={6}>
                    <Autocomplete
                      id="region"
                      style={{ width: 260 }} // fullWidth // doesn't exist on AutocompleteProps
                      options={USStates}
                      classes={{option: classes.option,}}
                      className={classes.selectField}
                      autoHighlight
                      getOptionLabel={option => option.label}
                      renderOption={option => (
                        <React.Fragment>
                          {option.label}
                        </React.Fragment>
                      )}
                      onChange={onRegionChange}
                      renderInput={params => (
                        <TextField
                          {...params}
                          label="Choose a state"
                          variant="outlined"
                          fullWidth
                          inputProps={{
                            ...params.inputProps,
                            autoComplete: 'new-password', // disable autocomplete and autofill
                          }}
                        />
                      )}
                    />
                  </Grid>
                )}
                <Grid item xs={12} sm={6}>
                  <TextField
                    id="company"
                    name="company"
                    label="Company Name"
                    fullWidth
                    value={values.companyName}
                    onChange={handleChange("companyName")}
                  />
                </Grid>
                <Grid item xs={12} sm={6}>
                  <TextField
                    id="taxID"
                    name="taxID"
                    label="Tax ID"
                    fullWidth
                    value={values.taxID}
                    onChange={handleChange("taxID")}
                  />
                  
                </Grid>
              </Grid>
            </Grid>
            <Grid item>
              <Typography className={classes.proratedDescription}>
                You will be charged a pro-rated amount for the current month. Receipts will be sent to your email each month.
              </Typography>
            </Grid>
          </Grid>
        </DialogContent>
        <DialogActions>
          <Button color="primary" autoFocus onClick={() => { closeDialogue() }}>
            Cancel
          </Button>
          <Button color="primary" variant="contained" autoFocus onClick={() => {
            if (!values.billingMethod) {
              setError("Please select your billing method.")
              setErrorDialogue(true)
            } else if (!values.country) {
              setError("Please select your country.")
              setErrorDialogue(true)
            } else if (values.country == "United States of America" && !values.region) {
              setError("Please select your state.")
              setErrorDialogue(true)
            } else if (values.companyName != "" && !values.taxID) {
              setError("Please provide your tax ID.")
              setErrorDialogue(true)
            } else {
              updateBillingInfo({
                variables: {
                  organizationID: organizationID,
                  billingMethodID: values.billingMethod,
                  billingPlanID: proPlan.billingPlanID,
                  country: values.country,
                  region: values.region,
                  companyName: values.companyName,
                  taxNumber: values.taxID
                }
              })
            }
          }}>
            Purchase
          </Button>
          <Dialog
            open={errorDialogue}
            aria-describedby="alert-dialog-description"
          >
            <DialogContent>
              {errorDialogue && (
                <Typography variant="body1" color="error">
                  {error}
                </Typography>)}
            </DialogContent>
            <DialogActions>
              <Button
                onClick={() => setErrorDialogue(false)}
                color="primary"
                autoFocus>
                Ok
              </Button>
            </DialogActions>
          </Dialog>
          <Dialog
            open={successDialogue}
            aria-describedby="alert-dialog-description"
          >
            <DialogContent>
                <Typography variant="body1">
                  Thank you for your purchase!
                </Typography>
            </DialogContent>
            <DialogActions>
              <Button
                onClick={() => {
                  setSuccessDialogue(false)
                  window.location.reload(true)
                }}
                color="primary"
                autoFocus>
                Ok
              </Button>
            </DialogActions>
          </Dialog>
        </DialogActions>
        {mutError && (
          <DialogContent>
            <Typography variant="body1" color="error">{mutError.message.replace("GraphQL error: ", "")}</Typography>
          </DialogContent>
        )}
      </Dialog>

      <Dialog
        open={route == "change_billing_method"}
        fullWidth={true}
        maxWidth={"sm"}
        onBackdropClick={() => { closeDialogue() }}
      >
        <DialogTitle id="alert-dialog-title">{"Change billing method"}</DialogTitle>
        <DialogContent>
          <Grid container direction="column">
            <Grid item>
              <SelectField
                id="billing_method"
                label="Billing method"
                value={values.billingMethod}
                options={data2.billingMethods.filter(billingMethod => billingMethod.paymentsDriver != billing.ANARCHISM_DRIVER).map((billingMethod) => {
                  if (billingMethod.paymentsDriver == billing.STRIPECARD_DRIVER) {
                    const payload = JSON.parse(billingMethod.driverPayload)
                    return { label: payload.brand.charAt(0).toUpperCase() + payload.brand.slice(1) + " xxxx-xxxx-xxxx-" + payload.last4, value: billingMethod.billingMethodID }
                  } else if (billingMethod.paymentsDriver == billing.STRIPEWIRE_DRIVER) {
                    return { label: "Wire payment", value: billingMethod.billingMethodID }
                  } else {
                    return { label: "", value: "" }
                  }
                })}
                onChange={({ target }) => setValues({ ...values, billingMethod: target.value as string })}
                controlClass={classes.selectField}
              />
            </Grid>
          </Grid>
        </DialogContent>
        <DialogActions>
          <Button color="primary" autoFocus onClick={() => {closeDialogue()}}>
            Cancel
          </Button>
          <Button color="primary" variant="contained" autoFocus onClick={() => {
            if (values.billingMethod) {
              updateBillingInfo({
                variables: {
                  organizationID: organizationID,
                  billingMethodID: values.billingMethod,
                  billingPlanID: proPlan.billingPlanID,
                  country: values.country
                }
              })
            } else {
              setError("Please select your billing method.")
              setErrorDialogue(true)
            }
          }}>
            Change Billing Method
          </Button>
          <Dialog
            open={errorDialogue}
            aria-describedby="alert-dialog-description"
          >
            <DialogContent>
              {errorDialogue && (
                <Typography variant="body1" color="error">
                  {error}
                </Typography>)}
            </DialogContent>
            <DialogActions>
              <Button
                onClick={() => setErrorDialogue(false)}
                color="primary"
                autoFocus>
                Ok
              </Button>
            </DialogActions>
          </Dialog>
        </DialogActions>
      </Dialog>
        
      <Dialog open={route=="cancel"}>
        <DialogTitle id="alert-dialog-title">{"Are you sure?"}</DialogTitle>
        <DialogContent>
          <DialogContentText id="alert-dialog-description">
            Upon canceling your plan, your usage will be assessed and you will be charged for any applicable overage fees for the current billing period.
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button color="primary" autoFocus onClick={() => closeDialogue()}>
            No, go back
          </Button>
          <Button color="primary" autoFocus onClick={() => {
            updateBillingInfo({ variables: { organizationID: organizationID, billingMethodID: anarchism.billingMethodID, billingPlanID: freePlan.billingPlanID, country: data.billingInfo.country } });
          }}>
            Yes, I'm sure
          </Button>
          <Dialog
            open={successDialogue}
            aria-describedby="alert-dialog-description"
          >
            <DialogContent>
              <Typography variant="body1">
                Your plan has been canceled.
              </Typography>
            </DialogContent>
            <DialogActions>
              <Button
                onClick={() => {
                  setSuccessDialogue(false)
                  window.location.reload(true)
                }}
                color="primary"
                autoFocus>
                Ok
              </Button>
            </DialogActions>
          </Dialog>
        </DialogActions>
      </Dialog>
    </>
  )
}

export default UpdateBillingInfoDialogue;

const countries = [
  { value: 'Afghanistan', label: 'Afghanistan' },
  { value: 'Åland Islands', label: 'Åland Islands' },
  { value: 'Albania', label: 'Albania' },
  { value: 'Algeria', label: 'Algeria' },
  { value: 'American Samoa', label: 'American Samoa' },
  { value: 'Andorra', label: 'Andorra' },
  { value: 'Angola', label: 'Angola' },
  { value: 'Anguilla', label: 'Anguilla' },
  { value: 'Antarctica', label: 'Antarctica' },
  { value: 'Antigua and Barbuda', label: 'Antigua and Barbuda' },
  { value: 'Argentina', label: 'Argentina' },
  { value: 'Armenia', label: 'Armenia' },
  { value: 'Aruba', label: 'Aruba' },
  { value: 'Australia', label: 'Australia' },
  { value: 'Austria', label: 'Austria' },
  { value: 'Azerbaijan', label: 'Azerbaijan' },
  { value: 'Bahamas', label: 'Bahamas' },
  { value: 'Bahrain', label: 'Bahrain' },
  { value: 'Bangladesh', label: 'Bangladesh' },
  { value: 'Barbados', label: 'Barbados' },
  { value: 'Belarus', label: 'Belarus' },
  { value: 'Belgium', label: 'Belgium' },
  { value: 'Belize', label: 'Belize' },
  { value: 'Benin', label: 'Benin' },
  { value: 'Bermuda', label: 'Bermuda' },
  { value: 'Bhutan', label: 'Bhutan' },
  { value: 'Bolivia (Plurinational State of)', label: 'Bolivia (Plurinational State of)' },
  { value: 'Bonaire, Sint Eustatius and Saba', label: 'Bonaire, Sint Eustatius and Saba' },
  { value: 'Bosnia and Herzegovina', label: 'Bosnia and Herzegovina' },
  { value: 'Botswana', label: 'Botswana' },
  { value: 'Bouvet Island', label: 'Bouvet Island' },
  { value: 'Brazil', label: 'Brazil' },
  { value: 'British Indian Ocean Territory', label: 'British Indian Ocean Territory' },
  { value: 'Brunei Darussalam', label: 'Brunei Darussalam' },
  { value: 'Bulgaria', label: 'Bulgaria' },
  { value: 'Burkina Faso', label: 'Burkina Faso' },
  { value: 'Burundi', label: 'Burundi' },
  { value: 'Cabo Verde', label: 'Cabo Verde' },
  { value: 'Cambodia', label: 'Cambodia' },
  { value: 'Cameroon', label: 'Cameroon' },
  { value: 'Canada', label: 'Canada' },
  { value: 'Cayman Islands', label: 'Cayman Islands' },
  { value: 'Central African Republic', label: 'Central African Republic' },
  { value: 'Chad', label: 'Chad' },
  { value: 'Chile', label: 'Chile' },
  { value: 'China', label: 'China' },
  { value: 'Christmas Island', label: 'Christmas Island' },
  { value: 'Cocos (Keeling) Islands', label: 'Cocos (Keeling) Islands' },
  { value: 'Colombia', label: 'Colombia' },
  { value: 'Comoros', label: 'Comoros' },
  { value: 'Congo', label: 'Congo' },
  { value: 'Congo, Democratic Republic of the', label: 'Congo, Democratic Republic of the' },
  { value: 'Cook Islands', label: 'Cook Islands' },
  { value: 'Costa Rica', label: 'Costa Rica' },
  { value: "Côte d'Ivoire", label: "Côte d'Ivoire" },
  { value: 'Croatia', label: 'Croatia' },
  { value: 'Cuba', label: 'Cuba' },
  { value: 'Curaçao', label: 'Curaçao' },
  { value: 'Cyprus', label: 'Cyprus' },
  { value: 'Czechia', label: 'Czechia' },
  { value: 'Denmark', label: 'Denmark' },
  { value: 'Djibouti', label: 'Djibouti' },
  { value: 'Dominica', label: 'Dominica' },
  { value: 'Dominican Republic', label: 'Dominican Republic' },
  { value: 'Ecuador', label: 'Ecuador' },
  { value: 'Egypt', label: 'Egypt' },
  { value: 'El Salvador', label: 'El Salvador' },
  { value: 'Equatorial Guinea', label: 'Equatorial Guinea' },
  { value: 'Eritrea', label: 'Eritrea' },
  { value: 'Estonia', label: 'Estonia' },
  { value: 'Eswatini', label: 'Eswatini' },
  { value: 'Ethiopia', label: 'Ethiopia' },
  { value: 'Falkland Islands (Malvinas)', label: 'Falkland Islands (Malvinas)' },
  { value: 'Faroe Islands', label: 'Faroe Islands' },
  { value: 'Fiji', label: 'Fiji' },
  { value: 'Finland', label: 'Finland' },
  { value: 'France', label: 'France' },
  { value: 'French Guiana', label: 'French Guiana' },
  { value: 'French Polynesia', label: 'French Polynesia' },
  { value: 'French Southern Territories', label: 'French Southern Territories' },
  { value: 'Gabon', label: 'Gabon' },
  { value: 'Gambia', label: 'Gambia' },
  { value: 'Georgia', label: 'Georgia' },
  { value: 'Germany', label: 'Germany' },
  { value: 'Ghana', label: 'Ghana' },
  { value: 'Gibraltar', label: 'Gibraltar' },
  { value: 'Greece', label: 'Greece' },
  { value: 'Greenland', label: 'Greenland' },
  { value: 'Grenada', label: 'Grenada' },
  { value: 'Guadeloupe', label: 'Guadeloupe' },
  { value: 'Guam', label: 'Guam' },
  { value: 'Guatemala', label: 'Guatemala' },
  { value: 'Guernsey', label: 'Guernsey' },
  { value: 'Guinea', label: 'Guinea' },
  { value: 'Guinea-Bissau', label: 'Guinea-Bissau' },
  { value: 'Guyana', label: 'Guyana' },
  { value: 'Haiti', label: 'Haiti' },
  { value: 'Heard Island and McDonald Islands', label: 'Heard Island and McDonald Islands' },
  { value: 'Holy See', label: 'Holy See' },
  { value: 'Honduras', label: 'Honduras' },
  { value: 'Hong Kong', label: 'Hong Kong' },
  { value: 'Hungary', label: 'Hungary' },
  { value: 'Iceland', label: 'Iceland' },
  { value: 'India', label: 'India' },
  { value: 'Indonesia', label: 'Indonesia' },
  { value: 'Iran (Islamic Republic of)', label: 'Iran (Islamic Republic of)' },
  { value: 'Iraq', label: 'Iraq' },
  { value: 'Ireland', label: 'Ireland' },
  { value: 'Isle of Man', label: 'Isle of Man' },
  { value: 'Israel', label: 'Israel' },
  { value: 'Italy', label: 'Italy' },
  { value: 'Jamaica', label: 'Jamaica' },
  { value: 'Japan', label: 'Japan' },
  { value: 'Jersey', label: 'Jersey' },
  { value: 'Jordan', label: 'Jordan' },
  { value: 'Kazakhstan', label: 'Kazakhstan' },
  { value: 'Kenya', label: 'Kenya' },
  { value: 'Kiribati', label: 'Kiribati' },
  { value: 'North Korea', label: 'North Korea' },
  { value: 'South Korea', label: 'South Korea' },
  { value: 'Kuwait', label: 'Kuwait' },
  { value: 'Kyrgyzstan', label: 'Kyrgyzstan' },
  { value: "Lao People's Democratic Republic", label: "Lao People's Democratic Republic" },
  { value: 'Latvia', label: 'Latvia' },
  { value: 'Lebanon', label: 'Lebanon' },
  { value: 'Lesotho', label: 'Lesotho' },
  { value: 'Liberia', label: 'Liberia' },
  { value: 'Libya', label: 'Libya' },
  { value: 'Liechtenstein', label: 'Liechtenstein' },
  { value: 'Lithuania', label: 'Lithuania' },
  { value: 'Luxembourg', label: 'Luxembourg' },
  { value: 'Macao', label: 'Macao' },
  { value: 'Madagascar', label: 'Madagascar' },
  { value: 'Malawi', label: 'Malawi' },
  { value: 'Malaysia', label: 'Malaysia' },
  { value: 'Maldives', label: 'Maldives' },
  { value: 'Mali', label: 'Mali' },
  { value: 'Malta', label: 'Malta' },
  { value: 'Marshall Islands', label: 'Marshall Islands' },
  { value: 'Martinique', label: 'Martinique' },
  { value: 'Mauritania', label: 'Mauritania' },
  { value: 'Mauritius', label: 'Mauritius' },
  { value: 'Mayotte', label: 'Mayotte' },
  { value: 'Mexico', label: 'Mexico' },
  { value: 'Micronesia (Federated States of)', label: 'Micronesia (Federated States of)' },
  { value: 'Moldova, Republic of', label: 'Moldova, Republic of' },
  { value: 'Monaco', label: 'Monaco' },
  { value: 'Mongolia', label: 'Mongolia' },
  { value: 'Montenegro', label: 'Montenegro' },
  { value: 'Montserrat', label: 'Montserrat' },
  { value: 'Morocco', label: 'Morocco' },
  { value: 'Mozambique', label: 'Mozambique' },
  { value: 'Myanmar', label: 'Myanmar' },
  { value: 'Namibia', label: 'Namibia' },
  { value: 'Nauru', label: 'Nauru' },
  { value: 'Nepal', label: 'Nepal' },
  { value: 'Netherlands', label: 'Netherlands' },
  { value: 'New Caledonia', label: 'New Caledonia' },
  { value: 'New Zealand', label: 'New Zealand' },
  { value: 'Nicaragua', label: 'Nicaragua' },
  { value: 'Niger', label: 'Niger' },
  { value: 'Nigeria', label: 'Nigeria' },
  { value: 'Niue', label: 'Niue' },
  { value: 'Norfolk Island', label: 'Norfolk Island' },
  { value: 'North Macedonia', label: 'North Macedonia' },
  { value: 'Northern Mariana Islands', label: 'Northern Mariana Islands' },
  { value: 'Norway', label: 'Norway' },
  { value: 'Oman', label: 'Oman' },
  { value: 'Pakistan', label: 'Pakistan' },
  { value: 'Palau', label: 'Palau' },
  { value: 'Palestine, State of', label: 'Palestine, State of' },
  { value: 'Panama', label: 'Panama' },
  { value: 'Papua New Guinea', label: 'Papua New Guinea' },
  { value: 'Paraguay', label: 'Paraguay' },
  { value: 'Peru', label: 'Peru' },
  { value: 'Philippines', label: 'Philippines' },
  { value: 'Pitcairn', label: 'Pitcairn' },
  { value: 'Poland', label: 'Poland' },
  { value: 'Portugal', label: 'Portugal' },
  { value: 'Puerto Rico', label: 'Puerto Rico' },
  { value: 'Qatar', label: 'Qatar' },
  { value: 'Réunion', label: 'Réunion' },
  { value: 'Romania', label: 'Romania' },
  { value: 'Russian Federation', label: 'Russian Federation' },
  { value: 'Rwanda', label: 'Rwanda' },
  { value: 'Saint Barthélemy', label: 'Saint Barthélemy' },
  { value: 'Saint Helena, Ascension and Tristan da Cunha', label: 'Saint Helena, Ascension and Tristan da Cunha' },
  { value: 'Saint Kitts and Nevis', label: 'Saint Kitts and Nevis' },
  { value: 'Saint Lucia', label: 'Saint Lucia' },
  { value: 'Saint Martin (French part)', label: 'Saint Martin (French part)' },
  { value: 'Saint Pierre and Miquelon', label: 'Saint Pierre and Miquelon' },
  { value: 'Saint Vincent and the Grenadines', label: 'Saint Vincent and the Grenadines' },
  { value: 'Samoa', label: 'Samoa' },
  { value: 'San Marino', label: 'San Marino' },
  { value: 'Sao Tome and Principe', label: 'Sao Tome and Principe' },
  { value: 'Saudi Arabia', label: 'Saudi Arabia' },
  { value: 'Senegal', label: 'Senegal' },
  { value: 'Serbia', label: 'Serbia' },
  { value: 'Seychelles', label: 'Seychelles' },
  { value: 'Sierra Leone', label: 'Sierra Leone' },
  { value: 'Singapore', label: 'Singapore' },
  { value: 'Sint Maarten (Dutch part)', label: 'Sint Maarten (Dutch part)' },
  { value: 'Slovakia', label: 'Slovakia' },
  { value: 'Slovenia', label: 'Slovenia' },
  { value: 'Solomon Islands', label: 'Solomon Islands' },
  { value: 'Somalia', label: 'Somalia' },
  { value: 'South Africa', label: 'South Africa' },
  { value: 'South Georgia and the South Sandwich Islands', label: 'South Georgia and the South Sandwich Islands' },
  { value: 'South Sudan', label: 'South Sudan' },
  { value: 'Spain', label: 'Spain' },
  { value: 'Sri Lanka', label: 'Sri Lanka' },
  { value: 'Sudan', label: 'Sudan' },
  { value: 'Suriname', label: 'Suriname' },
  { value: 'Svalbard and Jan Mayen', label: 'Svalbard and Jan Mayen' },
  { value: 'Sweden', label: 'Sweden' },
  { value: 'Switzerland', label: 'Switzerland' },
  { value: 'Syrian Arab Republic', label: 'Syrian Arab Republic' },
  { value: 'Taiwan, Province of China', label: 'Taiwan, Province of China' },
  { value: 'Tajikistan', label: 'Tajikistan' },
  { value: 'Tanzania, United Republic of', label: 'Tanzania, United Republic of' },
  { value: 'Thailand', label: 'Thailand' },
  { value: 'Timor-Leste', label: 'Timor-Leste' },
  { value: 'Togo', label: 'Togo' },
  { value: 'Tokelau', label: 'Tokelau' },
  { value: 'Tonga', label: 'Tonga' },
  { value: 'Trinidad and Tobago', label: 'Trinidad and Tobago' },
  { value: 'Tunisia', label: 'Tunisia' },
  { value: 'Turkey', label: 'Turkey' },
  { value: 'Turkmenistan', label: 'Turkmenistan' },
  { value: 'Turks and Caicos Islands', label: 'Turks and Caicos Islands' },
  { value: 'Tuvalu', label: 'Tuvalu' },
  { value: 'Uganda', label: 'Uganda' },
  { value: 'Ukraine', label: 'Ukraine' },
  { value: 'United Arab Emirates', label: 'United Arab Emirates' },
  { value: 'United Kingdom of Great Britain and Northern Ireland', label: 'United Kingdom of Great Britain and Northern Ireland' },
  { value: 'United States Minor Outlying Islands', label: 'United States Minor Outlying Islands' },
  { value: 'United States of America', label: 'United States of America' },
  { value: 'Uruguay', label: 'Uruguay' },
  { value: 'Uzbekistan', label: 'Uzbekistan' },
  { value: 'Vanuatu', label: 'Vanuatu' },
  { value: 'Venezuela (Bolivarian Republic of)', label: 'Venezuela (Bolivarian Republic of)' },
  { value: 'Viet Nam', label: 'Viet Nam' },
  { value: 'Virgin Islands (British)', label: 'Virgin Islands (British)' },
  { value: 'Virgin Islands (U.S.)', label: 'Virgin Islands (U.S.)' },
  { value: 'Wallis and Futuna', label: 'Wallis and Futuna' },
  { value: 'Western Sahara', label: 'Western Sahara' },
  { value: 'Yemen', label: 'Yemen' },
  { value: 'Zambia', label: 'Zambia' },
  { value: 'Zimbabw', label: 'Zimbabwe' }
]

const USStates = [
  { label: 'Alabama', value: 'Alabama' },
  { label: 'Alaska', value: 'Alaska' },
  { label: 'Arizona', value: 'Arizona' },
  { label: 'Arkansas', value: 'Arkansas' },
  { label: 'California', value: 'California' },
  { label: 'Colorado', value: 'Colorado' },
  { label: 'Connecticut', value: 'Connecticut' },
  { label: 'Delaware', value: 'Delaware' },
  { label: 'District of Columbia', value: 'District of Columbia' },
  { label: 'Florida', value: 'Florida' },
  { label: 'Georgia', value: 'Georgia' },
  { label: 'Hawaii', value: 'Hawaii' },
  { label: 'Idaho', value: 'Idaho' },
  { label: 'Illinois', value: 'Illinois' },
  { label: 'Indiana', value: 'Indiana' },
  { label: 'Iowa', value: 'Iowa' },
  { label: 'Kansas', value: 'Kansas' },
  { label: 'Kentucky', value: 'Kentucky' },
  { label: 'Louisiana', value: 'Louisiana' },
  { label: 'Maine', value: 'Maine' },
  { label: 'Maryland', value: 'Maryland' },
  { label: 'Massachusetts', value: 'Massachusetts' },
  { label: 'Michigan', value: 'Michigan' },
  { label: 'Minnesota', value: 'Minnesota' },
  { label: 'Mississippi', value: 'Mississippi' },
  { label: 'Missouri', value: 'Missouri' },
  { label: 'Montana', value: 'Montana' },
  { label: 'Nebraska', value: 'Nebraska' },
  { label: 'Nevada', value: 'Nevada' },
  { label: 'New Hampshire', value: 'New Hampshire' },
  { label: 'New Jersey', value: 'New Jersey' },
  { label: 'New Mexico', value: 'New Mexico' },
  { label: 'New York', value: 'New York' },
  { label: 'North Carolina', value: 'North Carolina' },
  { label: 'North Dakota', value: 'North Dakota' },
  { label: 'Ohio', value: 'Ohio' },
  { label: 'Oklahoma', value: 'Oklahoma' },
  { label: 'Oregon', value: 'Oregon' },
  { label: 'Pennsylvania', value: 'Pennsylvania' },
  { label: 'Puerto Rico', value: 'Puerto Rico' },
  { label: 'Rhode Island', value: 'Rhode Island' },
  { label: 'South Carolina', value: 'South Carolina' },
  { label: 'South Dakota', value: 'South Dakota' },
  { label: 'Tennessee', value: 'Tennessee' },
  { label: 'Texas', value: 'Texas' },
  { label: 'Utah', value: 'Utah' },
  { label: 'Vermont', value: 'Vermont' },
  { label: 'Virginia', value: 'Virginia' },
  { label: 'Washington', value: 'Washington' },
  { label: 'West Virginia', value: 'West Virginia' },
  { label: 'Wisconsin', value: 'Wisconsin' },
  { label: 'Wyoming', value: 'Wyoming' }
]