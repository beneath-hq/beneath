import _ from "lodash";
import React, { FC } from "react";

import {
  Button,
  Checkbox,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  FormControl,
  FormControlLabel,
  Grid,
  Link,
  ListItem,
  Radio,
  RadioGroup,
  TextField,
  Typography,
} from "@material-ui/core";
import { makeStyles } from "@material-ui/core/styles";
import CheckIcon from "@material-ui/icons/Check";
import { Autocomplete } from "@material-ui/lab";
import SelectField from "../../SelectField";

import { useMutation, useQuery } from "@apollo/react-hooks";
import { UPDATE_BILLING_INFO } from "../../../apollo/queries/billinginfo";
import { QUERY_BILLING_METHODS } from "../../../apollo/queries/billingmethod";
import { QUERY_BILLING_PLANS } from "../../../apollo/queries/billingplan";
import { BillingInfo_billingInfo } from "../../../apollo/types/BillingInfo";
import { BillingMethods, BillingMethodsVariables } from "../../../apollo/types/BillingMethods";
import { BillingPlans } from "../../../apollo/types/BillingPlans";
import { OrganizationByName_organizationByName_PrivateOrganization } from "../../../apollo/types/OrganizationByName";
import { UpdateBillingInfo, UpdateBillingInfoVariables } from "../../../apollo/types/UpdateBillingInfo";
import billing from "../../../lib/billing";

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
    minWidth: 300,
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
  billingMethodID: string;
  country: string;
  region: string;
  taxEntity: string;
  companyName: string;
  taxID: string;
  acceptedTerms: boolean;
}

interface Props {
  organization: OrganizationByName_organizationByName_PrivateOrganization;
  route: string;
  closeDialogue: () => void;
  billingInfo: BillingInfo_billingInfo;
}

const UpdateBillingInfoDialogue: FC<Props> = ({ organization, route, closeDialogue, billingInfo }) => {
  const classes = useStyles();
  const [successDialogue, setSuccessDialogue] = React.useState(false);
  const [errorDialogue, setErrorDialogue] = React.useState(false);
  const [error, setError] = React.useState("");
  const [values, setValues] = React.useState<CheckoutStateTypes>({
    billingMethodID: "",
    country: "",
    region: "",
    taxEntity: "business",
    companyName: "",
    taxID: "",
    acceptedTerms: false,
  });

  const { loading, error: queryError, data } = useQuery<BillingPlans>(QUERY_BILLING_PLANS);

  const { loading: loading2, error: queryError2, data: data2 } =
    useQuery<BillingMethods, BillingMethodsVariables>(QUERY_BILLING_METHODS, {
      variables: { organizationID: organization.organizationID },
  });

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

  if (queryError || !data) {
    return <p>Error: {JSON.stringify(queryError)}</p>;
  }

  const freePlan = data.billingPlans.filter((billingPlan) => billingPlan.default)[0];
  const proPlan = data.billingPlans.filter((billingPlan) => !billingPlan.default && billingPlan.availableInUI)[0];

  if (queryError2 || !data2) {
    return <p>Error: {JSON.stringify(queryError2)}</p>;
  }

  let billingMethodOptions: any[] = [];
  if (data2.billingMethods && data2.billingMethods.length > 0) {
    billingMethodOptions = data2.billingMethods.map((billingMethod) => {
      if (billingMethod.paymentsDriver === billing.STRIPECARD_DRIVER ) {
        const payload = JSON.parse(billingMethod.driverPayload);
        return {
          label: payload.brand.charAt(0).toUpperCase() + payload.brand.slice(1) + " xxxx-xxxx-xxxx-" + payload.last4,
          value: billingMethod.billingMethodID };
      } else if (billingMethod.paymentsDriver === billing.STRIPEWIRE_DRIVER ) {
        return {
        label: "Wire payment",
        value: billingMethod.billingMethodID };
      } else if (billingMethod.paymentsDriver === billing.ANARCHISM_DRIVER ) {
        return {
          label: "Anarchy!",
          value: billingMethod.billingMethodID };
      } else {
        return { label: "Unrecognized payment driver.", value: "TODO" };
      }
    });
  }

  const handleChange = (name: string) => (event: any) => {
    if (name === "acceptedTerms") {
      setValues({ ...values, [name]: event.target.checked });
      return;
    }
    setValues({ ...values, [name]: event.target.value });
  };

  const handleAutocompleteChange = (name: string) => (object: any, value: any) => {
    if (value) {
      setValues({ ...values, [name]: value.value });
    }
  };

  const sanitize = (name: string, value: any) => {
    if (name === "region") {
      return (values.country === "United States of America") ? value : "";
    }
    if (name === "companyName") {
      return (values.taxEntity === "business") ? value : "";
    }
    if (name === "taxID") {
      return (values.taxEntity === "business") ? value : "";
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
                {[
                  "5 GB writes included in base. Then $2/GB.",
                  "25 GB reads included in base. Then $1/GB.",
                  "Private projects",
                  "Role-based access controls",
                ].map((feature) => {
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
                Tax information
              </Typography>
              <Grid container direction="column" spacing={3}>
                <Grid item>
                  <FormControl component="fieldset">
                    <RadioGroup row value={values.taxEntity} onChange={handleChange("taxEntity")}>
                      <FormControlLabel value="business" control={<Radio />} label="Business" />
                      <FormControlLabel value="individual" control={<Radio />} label="Individual" />
                    </RadioGroup>
                  </FormControl>
                </Grid>
                <Grid item container spacing={3}>
                  <Grid item xs={12} lg={4}>
                    <Autocomplete
                      id="country"
                      style={{ width: 260 }} // fullWidth // doesn't exist on AutocompleteProps
                      options={billing.COUNTRIES}
                      classes={{ option: classes.option }}
                      className={classes.selectField}
                      autoHighlight
                      getOptionLabel={(option) => option.label}
                      renderOption={(option) => <React.Fragment>{option.label}</React.Fragment>}
                      onChange={handleAutocompleteChange("country")}
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
                  {values.country !== "United States of America" && <Grid item xs={12} sm={4}></Grid>}
                  {values.country === "United States of America" && (
                    <Grid item xs={12} lg={4}>
                      <Autocomplete
                        id="region"
                        style={{ width: 260 }} // fullWidth // doesn't exist on AutocompleteProps
                        options={billing.US_STATES}
                        classes={{ option: classes.option }}
                        className={classes.selectField}
                        autoHighlight
                        getOptionLabel={(option) => option.label}
                        renderOption={(option) => <React.Fragment>{option.label}</React.Fragment>}
                        onChange={handleAutocompleteChange("region")}
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
                </Grid>
                <Grid item container spacing={3}>
                  {values.taxEntity === "business" && (
                    <>
                      <Grid item xs={12} sm={6} md={4}>
                        <TextField
                          id="company"
                          name="company"
                          label="Company Name"
                          fullWidth
                          value={values.companyName}
                          onChange={handleChange("companyName")}
                        />
                      </Grid>
                      <Grid item xs={12} sm={6} md={4}>
                        <TextField
                          id="taxID"
                          name="taxID"
                          label="Tax ID"
                          fullWidth
                          value={values.taxID}
                          onChange={handleChange("taxID")}
                        />
                      </Grid>
                    </>
                  )}
                </Grid>
              </Grid>
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
                value={values.billingMethodID}
                options={billingMethodOptions}
                onChange={({ target }) => setValues({ ...values, billingMethodID: target.value as string })}
                controlClass={classes.selectField}
              />
            </Grid>
            <Grid item>
              <Typography className={classes.proratedDescription}>
                You will be charged a pro-rated amount for the current month. Receipts will be sent to your email each
                month.
              </Typography>
              <FormControlLabel
                className={classes.proratedDescription}
                control={
                  <Checkbox
                    color="secondary"
                    name="terms"
                    checked={values.acceptedTerms}
                    onChange={handleChange("acceptedTerms")}
                  />
                }
                label={
                  <Typography>
                    I authorise Beneath to send instructions to the financial institution that issued my card to take
                    payments from my card account in accordance with the
                    <Link href="https://about.beneath.dev/enterprise"> terms </Link> of my agreement with you.
                  </Typography>
                }
              />
            </Grid>
          </Grid>
          <Grid container spacing={2} className={classes.button}>
            <Grid item>
              <Button
                color="primary"
                autoFocus
                onClick={() => {
                  closeDialogue();
                }}
              >
                Cancel
              </Button>
            </Grid>
            <Grid item>
              <Button
                color="primary"
                variant="outlined"
                autoFocus
                disabled={
                  !values.country ||
                  (values.country === "United States of America" && !values.region) ||
                  (values.taxEntity === "business" && !values.companyName) ||
                  (values.taxEntity === "business" && !values.taxID) ||
                  !values.billingMethodID ||
                  !values.acceptedTerms
                }
                onClick={() => {
                  updateBillingInfo({
                    variables: {
                      organizationID: organization.organizationID,
                      billingMethodID: values.billingMethodID,
                      billingPlanID: proPlan.billingPlanID,
                      country: values.country,
                      region: sanitize("region", values.region),
                      companyName: sanitize("companyName", values.companyName),
                      taxNumber: sanitize("taxID", values.taxID),
                    },
                  });
                }}
              >
                Purchase
              </Button>
              <Dialog open={errorDialogue} aria-describedby="alert-dialog-description">
                <DialogContent>
                  {errorDialogue && (
                    <Typography variant="body1" color="error">
                      {error}
                    </Typography>
                  )}
                </DialogContent>
                <DialogActions>
                  <Button onClick={() => setErrorDialogue(false)} color="primary" autoFocus>
                    Ok
                  </Button>
                </DialogActions>
              </Dialog>
              <Dialog open={successDialogue} aria-describedby="alert-dialog-description">
                <DialogContent>
                  <Typography variant="body1">Thank you for your purchase!</Typography>
                </DialogContent>
                <DialogActions>
                  <Button
                    onClick={() => {
                      setSuccessDialogue(false);
                      window.location.reload(true);
                    }}
                    color="primary"
                    autoFocus
                  >
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
                value={values.billingMethodID}
                options={billingMethodOptions}
                onChange={({ target }) => setValues({ ...values, billingMethodID: target.value as string })}
                controlClass={classes.selectField}
              />
            </Grid>
          </Grid>
          <Grid container spacing={2} className={classes.button}>
            <Grid item>
              <Button
                color="primary"
                autoFocus
                onClick={() => {
                  closeDialogue();
                }}
              >
                Cancel
              </Button>
            </Grid>
            <Grid item>
              <Button
                color="primary"
                variant="contained"
                autoFocus
                onClick={() => {
                  if (values.billingMethodID) {
                    updateBillingInfo({
                      variables: {
                        organizationID: organization.organizationID,
                        billingMethodID: values.billingMethodID,
                        billingPlanID: billingInfo.billingPlan.billingPlanID,
                        country: billingInfo.country,
                      },
                    });
                  } else {
                    setError("Please select your billing method.");
                    setErrorDialogue(true);
                  }
                }}
              >
                Change Billing Method
              </Button>
              {mutError && (
                <DialogContent>
                  <Typography variant="body1" color="error" className={classes.errorMsg}>
                    {mutError.message.replace("GraphQL error: ", "")}
                  </Typography>
                </DialogContent>
              )}
              <Dialog open={errorDialogue} aria-describedby="alert-dialog-description">
                <DialogContent>
                  {errorDialogue && (
                    <Typography variant="body1" color="error">
                      {error}
                    </Typography>
                  )}
                </DialogContent>
                <DialogActions>
                  <Button onClick={() => setErrorDialogue(false)} color="primary" autoFocus>
                    Ok
                  </Button>
                </DialogActions>
              </Dialog>
              <Dialog open={successDialogue} aria-describedby="alert-dialog-description">
                <DialogContent>
                  <Typography variant="body1">Success.</Typography>
                </DialogContent>
                <DialogActions>
                  <Button
                    onClick={() => {
                      setSuccessDialogue(false);
                      window.location.reload(true);
                    }}
                    color="primary"
                    autoFocus
                  >
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
              <Button
                color="primary"
                autoFocus
                onClick={() => {
                  updateBillingInfo({
                    variables: {
                      organizationID: organization.organizationID,
                      billingPlanID: freePlan.billingPlanID,
                      country: billingInfo.country,
                    },
                  });
                }}
              >
                Yes, I'm sure
              </Button>
              {mutError && (
                <DialogContent>
                  <Typography variant="body1" color="error" className={classes.errorMsg}>
                    {mutError.message.replace("GraphQL error: ", "")}
                  </Typography>
                </DialogContent>
              )}
              <Dialog open={successDialogue} aria-describedby="alert-dialog-description">
                <DialogContent>
                  <Typography variant="body1">Your plan has been canceled.</Typography>
                </DialogContent>
                <DialogActions>
                  <Button
                    onClick={() => {
                      setSuccessDialogue(false);
                      window.location.reload(true);
                    }}
                    color="primary"
                    autoFocus
                  >
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
