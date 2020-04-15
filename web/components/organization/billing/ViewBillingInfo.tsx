import React, { FC, useEffect } from 'react'
import { Button, Typography, Grid, Dialog, DialogActions, DialogTitle, DialogContent, DialogContentText, Select, InputLabel, MenuItem } from "@material-ui/core"
import { useQuery } from "@apollo/react-hooks";

import useMe from "../../../hooks/useMe";

import billing from "../../../lib/billing"
import { QUERY_BILLING_INFO } from '../../../apollo/queries/billinginfo';
import { BillingInfo, BillingInfoVariables, BillingInfo_billingInfo_billingMethod, BillingInfo_billingInfo } from '../../../apollo/types/BillingInfo';
import { UpdateBillingInfo, UpdateBillingInfoVariables } from '../../../apollo/types/UpdateBillingInfo';
import { UPDATE_BILLING_INFO } from "../../../apollo/queries/billinginfo";
import AnarchismDetails from "./driver/AnarchismDetails"
import CardDetails from "./driver/CardDetails"
import WireDetails from "./driver/WireDetails"
import ViewBillingMethods from "./ViewBillingMethods"
import { BillingMethods, BillingMethodsVariables } from '../../../apollo/types/BillingMethods';
import { QUERY_BILLING_METHODS } from '../../../apollo/queries/billingmethod';

import { useMutation } from "@apollo/react-hooks";
import { makeStyles } from "@material-ui/core/styles"
import connection from "../../../lib/connection"
import CardForm from './driver/CardForm';

const useStyles = makeStyles((theme) => ({
  title: {
    marginTop: theme.spacing(2),
    marginBottom: theme.spacing(2),
  },
  button: {
    marginTop: theme.spacing(3),
  },
  buttons: {
    marginTop: theme.spacing(3),
  },
}))

interface Props {
  organizationID: string
}

// 1. view billing method (edit)
// 2. view billing plan (edit)
// 3. TODO: create organization with Enterprise plan
const ViewBilling: FC<Props> = ({ organizationID }) => {  
  const classes = useStyles()
  const [cancelDialogue, setCancelDialogue] = React.useState(false)
  const [upgradeDialogue, setUpgradeDialogue] = React.useState(false)
  const [addCard, setAddCard] = React.useState(false)
  const [selectedCard, setSelectedCard] = React.useState("")
  const [cancel, setCancel] = React.useState(false)
  const [error, setError] = React.useState("")

  const me = useMe();
  if (!me) {
    return <p>Need to log in to see your current billing plan</p>
  }

  // fetch organization's billing info for the billing plan and payment driver
  const { loading, error: queryError, data } = useQuery<BillingInfo, BillingInfoVariables>(QUERY_BILLING_INFO, {
    variables: {
      organizationID: organizationID,
    },
  });

  const { loading: loading2, error: queryError2, data: data2 } = useQuery<BillingMethods, BillingMethodsVariables>(QUERY_BILLING_METHODS, {
    variables: {
      organizationID: organizationID,
    },
  });

  if (error || !data || !data2) {
    return <p>Error: {JSON.stringify(error)}</p>;
  }

  const cards = data2.billingMethods.filter(billingMethod => billingMethod.paymentsDriver == "stripecard")
  const wire = data2.billingMethods.filter(billingMethod => billingMethod.paymentsDriver == "stripewire")
  const anarchism = data2.billingMethods.filter(billingMethod => billingMethod.paymentsDriver == "anarchism")
  
  const [updateBillingInfo] = useMutation<UpdateBillingInfo, UpdateBillingInfoVariables>(UPDATE_BILLING_INFO, {
    onCompleted: (data) => {
      console.log(data)
      // TODO: handle errors
      // check error
      // if (!res.ok) {
      //   const json: any = await res.json()
      //   setError(json.error)
      // }
  
      // on success, reload the page
      window.location.reload(true)
    },
  })

  if (error || !data) {
    return <p>Error: {JSON.stringify(error)}</p>;
  }

  // if (data.billingInfo.billingPlan.period === '\u0002') {
  //   var billingPeriod: string = billing.MONTHLY_BILLING_PLAN_STRING
  // } else {
  //   return <p>Error: your organization has an unknown billing plan period</p>
  // }

  const displayBillingMethod = (billingInfo: BillingInfo_billingInfo ) => {
    if (billingInfo.billingPlan.description == billing.FREE_BILLING_PLAN_DESCRIPTION) {
      return "N/A"
    } else {
      return billingInfo.billingMethod.paymentsDriver
    }
  }

  const billingInfo = [
    { name: 'Plan name', detail: data.billingInfo.billingPlan.description },
    { name: 'Billing method', detail: displayBillingMethod(data.billingInfo) },
  ]
    
  return (
    <React.Fragment>
      <Typography>
        You can find information about our billing plans at about.beneath.dev/enterprise.
      </Typography>
      <Grid container>
        <Grid item xs={12} sm={6}>
          <Typography variant="h6" className={classes.title}>
            Billing methods
          </Typography>
          {cards.map(({ billingMethodID, driverPayload }) => (
            <CardDetails billingMethodID={billingMethodID} driverPayload={driverPayload} />
          ))}

          {wire.map(({ billingMethodID }) => (
            <WireDetails />
          ))}

          {cards.length == 0 && wire.length == 0 && (
            <Typography>
              You have no billing methods on file
            </Typography>
          )}
          <Button
            variant="contained"
            onClick={() => {setAddCard(true)}}
          >
            Add Credit Card
          </Button>
          <Dialog 
            open={addCard} 
            fullWidth={true} 
            maxWidth={"md"}
            onBackdropClick={() => {setAddCard(false)}}
          >
            <DialogTitle id="alert-dialog-title">{"Add a credit card"}</DialogTitle>
            <DialogContent>
              <CardForm />
            </DialogContent>
          </Dialog>
        </Grid>
        <Grid item container direction="column" xs={12} sm={6}>
          <Typography variant="h6" className={classes.title}>
            Billing info
          </Typography>
          <Grid container>
            {billingInfo.map(billingInfo => (
              <React.Fragment key={billingInfo.name}>
                <Grid item xs={6}>
                  <Typography gutterBottom>{billingInfo.name}</Typography>
                </Grid>
                <Grid item xs={6}>
                  <Typography gutterBottom>{billingInfo.detail}</Typography>
                </Grid>
              </React.Fragment>
            ))}
          </Grid>
          {data.billingInfo.billingPlan.description == billing.FREE_BILLING_PLAN_DESCRIPTION && (
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
                onBackdropClick={() => { setUpgradeDialogue(false) }}
              >
                <DialogTitle id="alert-dialog-title">{"Select your payment method"}</DialogTitle>
                <DialogContent>
                  <InputLabel id="label">Billing method</InputLabel>
                  <Select 
                    labelId="label" 
                    id="select"
                    onChange={(event: React.ChangeEvent<{ value: unknown }>) => setSelectedCard(event.target.value as string)}
                  >
                    {cards.map((card) => {
                      const payload = JSON.parse(card.driverPayload)
                      return <MenuItem value={card.billingMethodID}>{payload.brand} + {payload.last4}</MenuItem>
                    }
                    )}
                  </Select>
                  <DialogContentText id="alert-dialog-description">
                    
                  </DialogContentText>
                </DialogContent>
                <DialogActions>
                  <Button color="primary" autoFocus onClick={() => {
                    setUpgradeDialogue(false)
                  }}>
                    Cancel
                    </Button>
                  <Button color="primary" autoFocus onClick={() => {
                    updateBillingInfo({
                      variables: {
                        organizationID: organizationID,
                        billingMethodID: selectedCard,
                        billingPlanID: billing.PRO_MONTHLY_BILLING_PLAN_ID
                      }
                  })}}>
                    Purchase
                    </Button>
                </DialogActions>
              </Dialog>
            </Grid>
          )}
          <Grid item>
            <Button
              variant="contained"
              color="secondary"
              className={classes.button}
              onClick={() => {
                window.location.href = "https://docs.google.com/forms/d/e/1FAIpQLSdsO3kcT3yk0Cgc4MzkPR_d16jZiYQd7L0M3ZxGwdOYycGhIg/viewform?usp=sf_link"
              }}>
              Discuss Enterprise Plan
            </Button>
          </Grid>
          {data.billingInfo.billingPlan.description == billing.PRO_BILLING_PLAN_DESCRIPTION && (
            <Grid item container className={classes.buttons}>
              <Grid item>
                <Button
                  variant="outlined"
                  color="secondary"
                  size="small"
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
                      updateBillingInfo({ variables: { organizationID: organizationID, billingMethodID: anarchism[0].billingMethodID, billingPlanID: billing.FREE_BILLING_PLAN_ID } });
                    }}>
                      Yes, I'm sure
                  </Button>
                  </DialogActions>
                </Dialog>
              </Grid>
              <Grid item>
                <Button
                  variant="outlined"
                  color="primary"
                  className={classes.button}
                  onClick={() => {
                    // TODO: updateBillingInfoBillingPlan()
                  }}>
                  Change Payment Method
                </Button>
              </Grid>
            </Grid>)}
        </Grid>
      </Grid>
    </React.Fragment>
  )
};

export default ViewBilling;