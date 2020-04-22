import React, { FC } from 'react'
import { useQuery } from "@apollo/react-hooks";
import useMe from "../../../hooks/useMe";
import _ from 'lodash'

import { QUERY_BILLING_INFO } from '../../../apollo/queries/billinginfo';
import { QUERY_BILLING_PLANS } from '../../../apollo/queries/billingplan';
import { BillingInfo, BillingInfoVariables, BillingInfo_billingInfo } from '../../../apollo/types/BillingInfo';
import { BillingPlans } from '../../../apollo/types/BillingPlans';

import { Button, Typography, Grid, Dialog, DialogTitle, DialogContent } from "@material-ui/core"
import { makeStyles } from "@material-ui/core/styles"

import billing from "../../../lib/billing"
import UpdateBillingInfo from './UpdateBillingInfo';

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
  const [changeBillingMethodDialogue, setChangeBillingMethodDialogue] = React.useState(false)
  const [cancelDialogue, setCancelDialogue] = React.useState(false)

  const me = useMe();
  if (!me) {
    return <p>Need to log in to see your current billing plan</p>
  }

  const { loading, error: queryError1, data } = useQuery<BillingInfo, BillingInfoVariables>(QUERY_BILLING_INFO, {
    variables: {
      organizationID: organizationID,
    },
  });
  
  const { loading: loading3, error: queryError3, data: data3 } = useQuery<BillingPlans>(QUERY_BILLING_PLANS);

  if (queryError1 || !data) {
    return <p>Error: {JSON.stringify(queryError1)}</p>;
  }
  
  if (queryError3 || !data3) {
    return <p>Error: {JSON.stringify(queryError3)}</p>;
  }

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
    { name: 'Country', detail: data.billingInfo.country, editButton: false },
    { name: 'Region', detail: data.billingInfo.region, editButton: false },
    { name: 'Company Name', detail: data.billingInfo.companyName, editButton: false },
    { name: 'Tax ID', detail: data.billingInfo.taxNumber, editButton: false },
    { name: 'Billing method', detail: displayBillingMethod(data.billingInfo), editButton: true },
  ]

  const handleCloseDialogue = () => {
    if (upgradeDialogue) {
      setUpgradeDialogue(false)
    }
    if (changeBillingMethodDialogue) {
      setChangeBillingMethodDialogue(false)
    }
    if (cancelDialogue) {
      setCancelDialogue(false)
    }
    return 
  }

  if (upgradeDialogue) {
    return <UpdateBillingInfo organizationID={organizationID} route={"checkout"} closeDialogue={handleCloseDialogue} />
  }
  if (changeBillingMethodDialogue) {
    return <UpdateBillingInfo organizationID={organizationID} route={"change_billing_method"} closeDialogue={handleCloseDialogue} />
  }
  if (cancelDialogue) {
    return <UpdateBillingInfo organizationID={organizationID} route={"cancel"} closeDialogue={handleCloseDialogue} />
  }
    
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