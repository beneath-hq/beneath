import React, { FC } from 'react'
import { useQuery, useMutation } from "@apollo/react-hooks";
import useMe from "../../../hooks/useMe";
import _ from 'lodash'

import { QUERY_BILLING_INFO, UPDATE_BILLING_INFO } from '../../../apollo/queries/billinginfo';
import { QUERY_BILLING_METHODS } from '../../../apollo/queries/billingmethod';
import { QUERY_BILLING_PLANS } from '../../../apollo/queries/billingplan';
import { BillingInfo, BillingInfoVariables, BillingInfo_billingInfo } from '../../../apollo/types/BillingInfo';
import { UpdateBillingInfo, UpdateBillingInfoVariables } from '../../../apollo/types/UpdateBillingInfo';
import { BillingMethods, BillingMethodsVariables } from '../../../apollo/types/BillingMethods';
import { BillingPlans } from '../../../apollo/types/BillingPlans';

import { Button, Typography, Grid, Dialog, DialogActions, DialogTitle, DialogContent, DialogContentText, ListItem } from "@material-ui/core"
import SelectField from "../../SelectField";
import { makeStyles } from "@material-ui/core/styles"
import CheckIcon from '@material-ui/icons/Check';

import billing from "../../../lib/billing"

const useStyles = makeStyles((theme) => ({
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
  selectBillingMethodControl: {
    marginTop: theme.spacing(2),
    minWidth: 250,
  },
  proratedDescription: {
    marginTop: theme.spacing(3)
  },
}))

interface Props {
  organizationID: string
}

const ViewBillingInfo: FC<Props> = ({ organizationID }) => {  
  const classes = useStyles()
  const [upgradeDialogue, setUpgradeDialogue] = React.useState(false)
  const [errorDialogue, setErrorDialogue] = React.useState(false)
  const [error, setError] = React.useState("")
  const [changeBillingMethodDialogue, setChangeBillingMethodDialogue] = React.useState(false)
  const [cancelDialogue, setCancelDialogue] = React.useState(false)
  const [selectedBillingMethod, setSelectedBillingMethod] = React.useState("")

  const me = useMe();
  if (!me) {
    return <p>Need to log in to see your current billing plan</p>
  }

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

  const [updateBillingInfo] = useMutation<UpdateBillingInfo, UpdateBillingInfoVariables>(UPDATE_BILLING_INFO, {
    onCompleted: (data) => {
      window.location.reload(true)
    },
    onError: (error) => {
      setError(error.message.replace("GraphQL error:", ""))
      setErrorDialogue(true)
    },
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

  const displayBillingMethod = (billingInfo: BillingInfo_billingInfo) => {
    if (billingInfo.billingPlan.billingPlanID == freePlan.billingPlanID) {
      return "N/A"
    } else if (billingInfo.billingMethod.paymentsDriver == billing.STRIPECARD_DRIVER) {
      const payload = JSON.parse(billingInfo.billingMethod.driverPayload)
      return payload.brand.charAt(0).toUpperCase() + payload.brand.slice(1) + " " + payload.last4
    } else if (billingInfo.billingMethod.paymentsDriver == billing.STRIPEWIRE_DRIVER) {
      return "Wire"
    }
  }

  const billingInfo = [
    { name: 'Plan name', detail: data.billingInfo.billingPlan.description, editButton: false },
    { name: 'Read quota', detail: (data.billingInfo.billingPlan.seatReadQuota / 10 ** 9).toString() + " GB", editButton: false },
    { name: 'Write quota', detail: (data.billingInfo.billingPlan.seatWriteQuota / 10 ** 9).toString() + " GB", editButton: false },
    { name: 'Billing method', detail: displayBillingMethod(data.billingInfo), editButton: true },
  ]
    
  return (
    <React.Fragment>
      <Grid item container direction="column" xs={12} sm={6}>
        <Typography variant="h6" className={classes.title}>
          Billing info
        </Typography>
        <Grid container alignItems="center">
          {billingInfo.map(billingInfo => (
            <React.Fragment key={billingInfo.name}>
              <Grid item xs={6}>
                <Typography>{billingInfo.name}</Typography>
              </Grid>
              <Grid item>
                <Typography>{billingInfo.detail}</Typography>
              </Grid>
              {data.billingInfo.billingPlan.billingPlanID == proPlan.billingPlanID && billingInfo.editButton && (
                <Grid item>
                  <Button color="primary"
                    onClick={() => {
                      setChangeBillingMethodDialogue(true)
                    }}>
                    Edit
                  </Button>
                  <Dialog
                    open={changeBillingMethodDialogue}
                    fullWidth={true}
                    maxWidth={"sm"}
                    onBackdropClick={() => { setChangeBillingMethodDialogue(false) }}
                  >
                    <DialogTitle id="alert-dialog-title">{"Change billing method"}</DialogTitle>
                    <DialogContent>
                      <Grid container direction="column">
                        <Grid item>
                          <SelectField
                            id="billing_method"
                            label="Billing method"
                            value={selectedBillingMethod}
                            options={data2.billingMethods.filter(billingMethod => billingMethod.paymentsDriver != billing.ANARCHISM_DRIVER).map((billingMethod) => {
                              if (billingMethod.paymentsDriver == billing.STRIPECARD_DRIVER) {
                                const payload = JSON.parse(billingMethod.driverPayload)
                                return { label: payload.brand.charAt(0).toUpperCase() + payload.brand.slice(1) + " xxxx-xxxx-xxxx-" + payload.last4, value: billingMethod.billingMethodID }
                              } else if (billingMethod.paymentsDriver == billing.STRIPEWIRE_DRIVER) {
                                return { label: "Wire payment", value: billingMethod.billingMethodID }
                              } else {
                                return { label: "", value: ""}
                              }
                            })}
                            onChange={({ target }) => setSelectedBillingMethod(target.value as string)}
                            controlClass={classes.selectBillingMethodControl}
                          />
                        </Grid>
                      </Grid>
                    </DialogContent>
                    <DialogActions>
                      <Button color="primary" autoFocus onClick={() => {
                        setChangeBillingMethodDialogue(false)
                      }}>
                        Cancel
                      </Button>
                      <Button color="primary" variant="contained" autoFocus onClick={() => {
                        if (selectedBillingMethod) {
                          updateBillingInfo({
                            variables: {
                              organizationID: organizationID,
                              billingMethodID: selectedBillingMethod,
                              billingPlanID: proPlan.billingPlanID
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
                </Grid>
              )}
            </React.Fragment>
          ))}
        </Grid>
        {data.billingInfo.billingPlan.billingPlanID == freePlan.billingPlanID && (
          <Grid item>
            <Button
              variant="contained"
              color="primary"
              className={classes.button}
              onClick={() => {
                setUpgradeDialogue(true)
              }}>
              Upgrade to Professional Plan
            </Button>
            <Dialog 
              open={upgradeDialogue}
              fullWidth={true}
              maxWidth={"sm"}
              onBackdropClick={() => { setUpgradeDialogue(false) }}
            >
              <DialogTitle id="alert-dialog-title">{"Checkout"}</DialogTitle>
              <DialogContent>
                <Grid container direction="column">
                  <Grid item>
                    <Typography variant="h2" className={classes.title}>
                      Professional plan: $50/month base
                    </Typography>
                    <Typography>
                      {["5 GB writes included in base. Then $2/GB.", "25 GB reads included in base. Then $1/GB.", "Private projects", "Role-based access controls"].map((feature) => {
                        return (
                          <React.Fragment key={feature}>
                            <ListItem>
                              <CheckIcon className={classes.icon}/>
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
                      value={selectedBillingMethod}
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
                      onChange={({ target }) => setSelectedBillingMethod(target.value as string)}
                      controlClass={classes.selectBillingMethodControl}
                    />
                  </Grid>
                  <Grid item>
                    <Typography className={classes.proratedDescription}>
                      You will be charged a pro-rated amount for the current month. Receipts will be sent to your email each month.
                    </Typography>
                  </Grid>
                </Grid>
              </DialogContent>
              <DialogActions className={classes.button}>
                <Button color="primary" autoFocus onClick={() => {
                  setUpgradeDialogue(false)
                }}>
                  Cancel
                </Button>
                <Button color="primary" variant="contained" autoFocus onClick={() => {
                  if (selectedBillingMethod) {
                    updateBillingInfo({
                      variables: {
                        organizationID: organizationID,
                        billingMethodID: selectedBillingMethod,
                        billingPlanID: proPlan.billingPlanID
                      }
                    })
                  } else {
                    setError("Please select your billing method.")
                    setErrorDialogue(true)
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
              </DialogActions>
            </Dialog>
            <Grid item>
              <Button
                variant="contained"
                className={classes.button}
                href="https://docs.google.com/forms/d/e/1FAIpQLSdsO3kcT3yk0Cgc4MzkPR_d16jZiYQd7L0M3ZxGwdOYycGhIg/viewform?usp=sf_link"
              >
                Discuss Enterprise Plan
              </Button>
            </Grid>
          </Grid>
        )}
        
        {data.billingInfo.billingPlan.billingPlanID == proPlan.billingPlanID && (
          <Grid item container direction="column">
            <Grid container item>

              <Grid item>
                <Button
                  variant="outlined"
                  // color="secondary"
                  className={classes.button}
                  onClick={() => {
                    setCancelDialogue(true)
                  }}>
                  Cancel plan
              </Button>
                <Dialog open={cancelDialogue}>
                  <DialogTitle id="alert-dialog-title">{"Are you sure?"}</DialogTitle>
                  <DialogContent>
                    <DialogContentText id="alert-dialog-description">
                      Upon canceling your plan, your usage will be assessed and you will be charged for any applicable overage fees for the current billing period.
                    </DialogContentText>
                  </DialogContent>
                  <DialogActions>
                    <Button color="primary" autoFocus onClick={() => {
                      setCancelDialogue(false)
                    }}>
                      No, go back
                  </Button>
                    <Button color="primary" autoFocus onClick={() => {
                      updateBillingInfo({ variables: { organizationID: organizationID, billingMethodID: anarchism.billingMethodID, billingPlanID: freePlan.billingPlanID } });
                    }}>
                      Yes, I'm sure
                  </Button>
                  </DialogActions>
                </Dialog>
              </Grid>
              <Grid item>
                <Button
                  variant="contained"
                  color="primary"
                  className={classes.button}
                  onClick={() => {
                    window.location.href = "https://docs.google.com/forms/d/e/1FAIpQLSdsO3kcT3yk0Cgc4MzkPR_d16jZiYQd7L0M3ZxGwdOYycGhIg/viewform?usp=sf_link"
                  }}>
                  Discuss Enterprise Plan
                </Button>
              </Grid>
            </Grid>
          </Grid>)}
      </Grid>
    </React.Fragment>
  )
};

export default ViewBillingInfo;