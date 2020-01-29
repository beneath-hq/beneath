import React, { FC, useEffect } from 'react'
import { Button, Typography } from "@material-ui/core"
import Grid from '@material-ui/core/Grid'
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
}))

interface Props {
  description: string | null
  billing_period: string
}

const CurrentBillingPlan: FC<Props> = ({ description, billing_period }) => {
  const [cancelPlan, setCancelPlan] = React.useState(false)
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
    const cancelProPlan = (async () => {
      if (cancelPlan) {
        const headers = { authorization: `Bearer ${token}` }
        let url = `${connection.API_URL}/billing/anarchism/initialize_customer`
        url += `?organizationID=${me.organization.organizationID}`
        url += `&billingPlanID=${billing.FREE_BILLING_PLAN_ID}`
        
        const res = await fetch(url, { headers })
        const json: any = await res.json()
        console.log(json)
        // if error, display error
        // if (!res.ok) {
        //   setValues({ ...values, ...{ error: json.error } })
        //   return
        // }

        // if success, say we're sorry to see you go, and provide link to refresh billing tab
      }
    })
    
    cancelProPlan()

  }, [cancelPlan])



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
        <Grid item>
          <Button
            variant="outlined"
            color="secondary"
            size="small"
            onClick={() => {
              // TODO: put this in a dialogue: are you sure?
              setCancelPlan(true)
            }}>
            Cancel plan
          </Button>
        </Grid>)}
      </Grid>
    </React.Fragment>
  )
}

export default CurrentBillingPlan;