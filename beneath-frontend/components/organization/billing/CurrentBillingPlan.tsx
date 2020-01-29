import React, { FC, useEffect } from 'react'
import { Button, Typography, Grid, Dialog, DialogActions, DialogTitle, DialogContent, DialogContentText } from "@material-ui/core"
import { makeStyles } from "@material-ui/core/styles"
import { useToken } from '../../../hooks/useToken'
import connection from "../../../lib/connection"
import billing from "../../../lib/billing"
import useMe from "../../../hooks/useMe";

const useStyles = makeStyles((theme) => ({
  title: {
    marginTop: theme.spacing(2),
    marginBottom: theme.spacing(2),
  },
  buttons: {
    marginTop: theme.spacing(3),
  },
}))

interface Props {
  description: string | null
  billing_period: string
}

const CurrentBillingPlan: FC<Props> = ({ description, billing_period }) => {
  const [openDialogue, setOpenDialogue] = React.useState(false)
  const [cancel, setCancel] = React.useState(false)
  const [error, setError] = React.useState("")
  const classes = useStyles()
  const token = useToken()

  const me = useMe(); // Q: is this in apollo local state?
  if (!me) {
    return <p>Need to log in to see your current billing plan</p>
  }

  const planDetails = [
    { name: 'Plan name', detail: description },
    { name: 'Billing cycle', detail: billing_period },
  ]

  useEffect(() => {
    const cancelPlan = (async () => {
      if (cancel) {
        const headers = { authorization: `Bearer ${token}` }
        let url = `${connection.API_URL}/billing/anarchism/initialize_customer`
        url += `?organizationID=${me.organization.organizationID}`
        url += `&billingPlanID=${billing.FREE_BILLING_PLAN_ID}`
        
        const res = await fetch(url, { headers })
        // check error
        if (!res.ok) {
          const json: any = await res.json()
          setError(json.error)
        }

        // on success, reload the page
        if (res.ok) {
          window.location.reload(true)
        }
      }
    })
    
    cancelPlan()

  }, [cancel])

  if (error) {
    return <p>{error}</p>
  }

  return (
    <React.Fragment>
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
        {description == billing.PRO_BILLING_PLAN_DESCRIPTION && (
        <Grid item container className={classes.buttons}>
          <Grid item>
            <Button
              variant="outlined"
              color="secondary"
              size="small"
              onClick={() => {
                setOpenDialogue(true)
              }}>
              Cancel plan
            </Button>
            <Dialog open={openDialogue}>
              <DialogTitle id="alert-dialog-title">{"Are you sure?"}</DialogTitle>
                <DialogContent>
                  <DialogContentText id="alert-dialog-description">
                    Upon canceling your plan, your usage will be assessed and you will be charged for any applicable overage fees for the current billing period. 
                  </DialogContentText>
                </DialogContent>
              <DialogActions>
                <Button color="primary" autoFocus onClick={() => {
                  setOpenDialogue(false)
                }}>
                  No, go back
                </Button>
                <Button color="primary" autoFocus onClick={() => {
                  setCancel(true)
                }}>
                  Yes, I'm sure
                </Button>
              </DialogActions>
            </Dialog>
          </Grid>
        </Grid>)}
      </Grid>
    </React.Fragment>
  )
}

export default CurrentBillingPlan;