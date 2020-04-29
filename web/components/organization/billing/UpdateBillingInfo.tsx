import _ from "lodash";
import React, { FC } from "react";

import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  Grid,
  ListItem,
  TextField,
  Typography,
} from "@material-ui/core";
import { makeStyles } from "@material-ui/core/styles";
import { Autocomplete } from "@material-ui/lab";

import CheckIcon from "@material-ui/icons/Check";

import billing from "../../../lib/billing";
import SelectField from "../../SelectField";

import { useMutation, useQuery } from "@apollo/react-hooks";
import { QUERY_BILLING_INFO, UPDATE_BILLING_INFO } from "../../../apollo/queries/billinginfo";
import { QUERY_BILLING_METHODS } from "../../../apollo/queries/billingmethod";
import { QUERY_BILLING_PLANS } from "../../../apollo/queries/billingplan";
import { BillingInfo, BillingInfoVariables } from "../../../apollo/types/BillingInfo";
import { BillingMethods, BillingMethodsVariables } from "../../../apollo/types/BillingMethods";
import { BillingPlans } from "../../../apollo/types/BillingPlans";
import { UpdateBillingInfo, UpdateBillingInfoVariables } from "../../../apollo/types/UpdateBillingInfo";

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
    "& > span": {
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
  errorMsg: {
    marginTop: theme.spacing(3)
  },
}));

interface CheckoutStateTypes {
  billingMethod: string;
  country: string;
  region: string;
  companyName: string;
  taxID: string;
}

interface Props {
  organizationID: string;
  route: string;
  closeDialogue: () => void;
}

const UpdateBillingInfoDialogue: FC<Props> = ({ organizationID, route, closeDialogue }) => {
  const classes = useStyles();
  const [successDialogue, setSuccessDialogue] = React.useState(false);
  const [errorDialogue, setErrorDialogue] = React.useState(false);
  const [error, setError] = React.useState("");
  const [values, setValues] = React.useState<CheckoutStateTypes>({
    billingMethod: "",
    country: "",
    region: "",
    companyName: "",
    taxID: "",
  });

  const { loading, error: queryError1, data } = useQuery<BillingInfo, BillingInfoVariables>(QUERY_BILLING_INFO, {
    variables: {organizationID},
  });

  const { loading: loading2, error: queryError2, data: data2 } =
    useQuery<BillingMethods, BillingMethodsVariables>(QUERY_BILLING_METHODS, {
    variables: {organizationID},
  });

  const { loading: loading3, error: queryError3, data: data3 } = useQuery<BillingPlans>(QUERY_BILLING_PLANS);

  const [updateBillingInfo, {error: mutError}] =
    useMutation<UpdateBillingInfo, UpdateBillingInfoVariables>(UPDATE_BILLING_INFO, {
    onCompleted: (data) => {
      if (data) {
        setSuccessDialogue(true);
      }
    },
    // this doesn't seem to work, but it'd be nice if it did!
    // onError: (error) => {
    //   console.log(error)
    //   setError(error.message.replace("GraphQL error:", ""))
    //   setErrorDialogue(true)
    // },
  });

  if (queryError1 || !data) {
    return <p>Error: {JSON.stringify(queryError1)}</p>;
  }

  if (queryError2 || !data2) {
    return <p>Error: {JSON.stringify(queryError2)}</p>;
  }

  if (queryError3 || !data3) {
    return <p>Error: {JSON.stringify(queryError3)}</p>;
  }

  // THURSDAY: FROM THIS
  // const cards = (data2.billingMethods ? data2.billingMethods.filter(billingMethod => billingMethod.paymentsDriver == billing.STRIPECARD_DRIVER).map((billingMethod) => {
  //   const payload = JSON.parse(billingMethod.driverPayload)
  //   return { label: payload.brand.charAt(0).toUpperCase() + payload.brand.slice(1) + " xxxx-xxxx-xxxx-" + payload.last4, value: billingMethod.billingMethodID }
  // }) : [])
  // const wire = data2.billingMethods.filter(billingMethod => billingMethod.paymentsDriver == billing.STRIPEWIRE_DRIVER).map((billingMethod) => {
  //   return { label: "Wire payment", value: billingMethod.billingMethodID }
  // })
  // const anarchism = data2.billingMethods.filter(billingMethod => billingMethod.paymentsDriver == billing.ANARCHISM_DRIVER)[0]
  // const billingMethodOptions = cards.concat(wire)

  // TO THIS
  // let cards;
  // let wire;
  // if (data2.billingMethods && data2.billingMethods.length > 0) {
  //   cards = data2.billingMethods.filter((billingMethod) => billingMethod.paymentsDriver === billing.STRIPECARD_DRIVER);
  //   wire = data2.billingMethods.filter((billingMethod) => billingMethod.paymentsDriver === billing.STRIPEWIRE_DRIVER)[0];
  // }

  const billingMethodOptions = [{ label: "lala", value: "tlja"}];

  const freePlan = data3.billingPlans.filter((billingPlan) => billingPlan.default)[0];
  const proPlan = data3.billingPlans.filter((billingPlan) => !billingPlan.default)[0];

  const handleChange = (name: string) => (event: any) => {
    setValues({ ...values, [name]: event.target.value });
  };

  const onCountryChange = (object: any, value: any) => {
    if (value) {
      setValues({ ...values, country: value.value });
    }
  };

  const onRegionChange = (object: any, value: any) => {
    if (value) {
      setValues({ ...values, region: value.value });
    }
  };

  return (
    <>
      {route === "checkout" && (
        <>
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
                  );
                })}
              </Typography>
            </Grid>
            <Grid item>
              <Typography variant="h2" className={classes.title}>
                Select your billing method
              </Typography>
              {billingMethodOptions.length === 0 && (
                <Typography variant="body1" color="error">
                  You have no billing method on file. Please add a card on the previous screen.
                </Typography>
              )}
              <SelectField
                id="billing_method"
                label="Billing method"
                value={values.billingMethod}
                options={billingMethodOptions}
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
                    options={billing.COUNTRIES}
                    classes={{option: classes.option, }}
                    className={classes.selectField}
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
                      />
                    )}
                  />
                </Grid>
                {values.country !== "United States of America" && (
                  <Grid item xs={12} sm={6}>
                  </Grid>
                )}
                {values.country === "United States of America" && (
                  <Grid item xs={12} sm={6}>
                    <Autocomplete
                      id="region"
                      style={{ width: 260 }} // fullWidth // doesn't exist on AutocompleteProps
                      options={billing.US_STATES}
                      classes={{option: classes.option, }}
                      className={classes.selectField}
                      autoHighlight
                      getOptionLabel={(option) => option.label}
                      renderOption={(option) => (
                        <React.Fragment>
                          {option.label}
                        </React.Fragment>
                      )}
                      onChange={onRegionChange}
                      renderInput={(params) => (
                        <TextField
                          {...params}
                          label="Choose a state"
                          variant="outlined"
                          fullWidth
                          inputProps={{
                            ...params.inputProps,
                            autoComplete: "new-password", // disable autocomplete and autofill
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
                You will be charged a pro-rated amount for the current month.
                Receipts will be sent to your email each month.
              </Typography>
            </Grid>
          </Grid>
          <Grid container spacing={2} className={classes.button}>
            <Grid item>
              <Button color="primary" autoFocus onClick={() => { closeDialogue(); }}>
                Cancel
              </Button>
            </Grid>
            <Grid item>
              <Button color="primary" variant="contained" autoFocus onClick={() => {
                if (!values.billingMethod) {
                  setError("Please select your billing method.");
                  setErrorDialogue(true);
                } else if (!values.country) {
                  setError("Please select your country.");
                  setErrorDialogue(true);
                } else if (values.country === "United States of America" && !values.region) {
                  setError("Please select your state.");
                  setErrorDialogue(true);
                } else if (values.companyName !== "" && !values.taxID) {
                  setError("Please provide your tax ID.");
                  setErrorDialogue(true);
                } else {
                  updateBillingInfo({
                    variables: {
                      organizationID,
                      billingMethodID: values.billingMethod,
                      billingPlanID: proPlan.billingPlanID,
                      country: values.country,
                      region: values.region,
                      companyName: values.companyName,
                      taxNumber: values.taxID
                    }
                  });
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
                      setSuccessDialogue(false);
                      window.location.reload(true);
                    }}
                    color="primary"
                    autoFocus>
                    Ok
                  </Button>
                </DialogActions>
              </Dialog>
            </Grid>
          </Grid>
          {mutError && (
            <DialogContent>
              <Typography variant="body1" color="error" className={classes.errorMsg}>
                {mutError.message.replace("GraphQL error: ", "")}
              </Typography>
            </DialogContent>
          )}
        </>
      )}

      {route === "change_billing_method" && (
        <>
          <Grid container direction="column">
            <Grid item>
              <SelectField
                id="billing_method"
                label="Billing method"
                value={values.billingMethod}
                options={billingMethodOptions}
                onChange={({ target }) => setValues({ ...values, billingMethod: target.value as string })}
                controlClass={classes.selectField}
              />
            </Grid>
          </Grid>
          <Grid container spacing={2} className={classes.button}>
            <Grid item>
              <Button color="primary" autoFocus onClick={() => {closeDialogue(); }}>
                Cancel
              </Button>
            </Grid>
            <Grid item>
              <Button color="primary" variant="contained" autoFocus onClick={() => {
                if (values.billingMethod) {
                  updateBillingInfo({
                    variables: {
                      organizationID,
                      billingMethodID: values.billingMethod,
                      billingPlanID: proPlan.billingPlanID,
                      country: data.billingInfo.country
                    }
                  });
                } else {
                  setError("Please select your billing method.");
                  setErrorDialogue(true);
                }
              }}>
                Change Billing Method
              </Button>
              {mutError && (
                <DialogContent>
                  <Typography variant="body1" color="error" className={classes.errorMsg}>
                    {mutError.message.replace("GraphQL error: ", "")}
                  </Typography>
                </DialogContent>
              )}
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
                    Success.
                    </Typography>
                </DialogContent>
                <DialogActions>
                  <Button
                    onClick={() => {
                      setSuccessDialogue(false);
                      window.location.reload(true);
                    }}
                    color="primary"
                    autoFocus>
                    Ok
                  </Button>
                </DialogActions>
              </Dialog>
            </Grid>
          </Grid>
        </>
      )}

      {route === "cancel" && (
        <>
          <DialogContentText id="alert-dialog-description">
            Upon canceling your plan, your usage will be assessed and you will be charged for any applicable overage
            fees for the current billing period.
          </DialogContentText>
          <Grid container spacing={2} className={classes.button}>
            <Grid item>
              <Button color="primary" autoFocus onClick={() => closeDialogue()}>
                No, go back
              </Button>
            </Grid>
            <Grid item>
              <Button color="primary" autoFocus onClick={() => {
                updateBillingInfo({ variables:
                  { organizationID,
                    billingMethodID: "",
                    billingPlanID: freePlan.billingPlanID,
                    country: data.billingInfo.country } });
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
                      setSuccessDialogue(false);
                      window.location.reload(true);
                    }}
                    color="primary"
                    autoFocus>
                    Ok
                  </Button>
                </DialogActions>
              </Dialog>
            </Grid>
          </Grid>
        </>
      )}
    </>
  ); };

export default UpdateBillingInfoDialogue;
