import { useQuery } from "@apollo/react-hooks";
import _ from "lodash";
import React, { FC } from "react";
import useMe from "../../../hooks/useMe";

import { QUERY_BILLING_INFO } from "../../../apollo/queries/billinginfo";
import { QUERY_BILLING_PLANS } from "../../../apollo/queries/billingplan";
import { BillingInfo, BillingInfo_billingInfo, BillingInfoVariables } from "../../../apollo/types/BillingInfo";
import { BillingPlans } from "../../../apollo/types/BillingPlans";

import { Button, Dialog, DialogActions, DialogContent, DialogTitle, Grid, Typography } from "@material-ui/core";
import { makeStyles } from "@material-ui/core/styles";

import billing from "../../../lib/billing";
import UpdateBillingInfo from "./UpdateBillingInfo";

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
}));

interface Props {
  organizationID: string;
}

const ViewBillingInfo: FC<Props> = ({ organizationID }) => {
  const classes = useStyles();
  const [upgradeDialogue, setUpgradeDialogue] = React.useState(false);
  const [changeBillingMethodDialogue, setChangeBillingMethodDialogue] = React.useState(false);
  const [cancelDialogue, setCancelDialogue] = React.useState(false);

  const me = useMe();
  if (!me) {
    return <p>Need to log in to see your current billing plan</p>;
  }

  const { loading, error: queryError1, data } = useQuery<BillingInfo, BillingInfoVariables>(QUERY_BILLING_INFO, {
    variables: {
      organizationID,
    },
  });

  const { loading: loading3, error: queryError3, data: data3 } = useQuery<BillingPlans>(QUERY_BILLING_PLANS);

  if (queryError1 || !data) {
    return <p>Error: {JSON.stringify(queryError1)}</p>;
  }

  if (queryError3 || !data3) {
    return <p>Error: {JSON.stringify(queryError3)}</p>;
  }

  const freePlan = data3.billingPlans.filter((billingPlan) => billingPlan.default)[0];
  const proPlan = data3.billingPlans.filter((billingPlan) => !billingPlan.default)[0];

  const displayBillingMethod = (billingInfo: BillingInfo_billingInfo) => {
    if (!billingInfo.billingMethod) {
      return "N/A";
    } else if (billingInfo.billingMethod.paymentsDriver === billing.STRIPECARD_DRIVER) {
      const payload = JSON.parse(billingInfo.billingMethod.driverPayload);
      return payload.brand.charAt(0).toUpperCase() + payload.brand.slice(1) + " " + payload.last4;
    } else if (billingInfo.billingMethod.paymentsDriver === billing.STRIPEWIRE_DRIVER) {
      return "Wire";
    }
  };

  const billingPlan = [
    { name: "Plan name", detail: data.billingInfo.billingPlan.description },
    { name: "Read quota", detail: (data.billingInfo.billingPlan.seatReadQuota / 10 ** 9).toString() + " GB" },
    { name: "Write quota", detail: (data.billingInfo.billingPlan.seatWriteQuota / 10 ** 9).toString() + " GB" },
  ];

  const taxInfo = [
    { name: "Country", detail: data.billingInfo.country },
    { name: "Region", detail: data.billingInfo.region },
    { name: "Company", detail: data.billingInfo.companyName },
    { name: "Tax ID", detail: data.billingInfo.taxNumber },
  ];

  const handleCloseDialogue = () => {
    if (upgradeDialogue) {
      setUpgradeDialogue(false);
    }
    if (changeBillingMethodDialogue) {
      setChangeBillingMethodDialogue(false);
    }
    if (cancelDialogue) {
      setCancelDialogue(false);
    }
    return;
  };

  return (
    <React.Fragment>
      <Grid container spacing={3}>
        <Grid item xs={12} md={4}>
          <Grid container>
            <Grid item>
              <Typography variant="h6" className={classes.title}>
                Billing plan
              </Typography>
            </Grid>
          </Grid>
          {billingPlan.map((billingPlan) => (
            <React.Fragment key={billingPlan.name}>
              <Grid container>
                <Grid item xs={6} sm={4}>
                  <Typography>{billingPlan.name}</Typography>
                </Grid>
                <Grid item>
                  <Typography>{billingPlan.detail}</Typography>
                </Grid>
              </Grid>
            </React.Fragment>
          ))}
        </Grid>
        <Grid item xs={12} md={4}>
          <Grid container alignItems="center" spacing={1}>
            <Grid item>
              <Typography variant="h6" className={classes.title}>
                Billing method
              </Typography>
            </Grid>
            {data.billingInfo.billingPlan.billingPlanID === proPlan.billingPlanID && (
              <Grid item>
                <Button
                  color="primary"
                  onClick={() => {setChangeBillingMethodDialogue(true); }}
                  className={classes.title}
                >
                  Edit
                </Button>
                <Dialog
                  open={changeBillingMethodDialogue}
                  fullWidth={true}
                  maxWidth={"sm"}
                  onBackdropClick={() => { setChangeBillingMethodDialogue(false); }}
                >
                  <DialogTitle id="alert-dialog-title">{"Change billing method"}</DialogTitle>
                  <DialogContent>
                    <UpdateBillingInfo
                    organizationID={organizationID}
                    route={"change_billing_method"}
                    closeDialogue={handleCloseDialogue} />
                  </DialogContent>
                  <DialogActions />
                </Dialog>
              </Grid>
            )}
          </Grid>
          <Typography>
            {displayBillingMethod(data.billingInfo)}
          </Typography>
        </Grid>
        <Grid item xs={12} md={4}>
          <Grid container>
            <Grid item>
              <Typography variant="h6" className={classes.title}>
                Tax info
              </Typography>
            </Grid>
          </Grid>
          {data.billingInfo.billingPlan.billingPlanID === freePlan.billingPlanID && (
            <Typography>N/A</Typography>
          )}
          {data.billingInfo.billingPlan.billingPlanID !== freePlan.billingPlanID && taxInfo.map((taxInfo) => (
            <React.Fragment key={taxInfo.name}>
              <Grid container>
                <Grid item xs={6} sm={4}>
                  <Typography>{taxInfo.name}</Typography>
                </Grid>
                <Grid item>
                  <Typography>{taxInfo.detail}</Typography>
                </Grid>
              </Grid>
            </React.Fragment>
          ))}
        </Grid>
      </Grid>

      {/* BUTTONS */}
      {data.billingInfo.billingPlan.billingPlanID === freePlan.billingPlanID && (
        <Grid container>
          <Grid item>
            <Button
              variant="contained"
              color="primary"
              className={classes.button}
              onClick={() => {setUpgradeDialogue(true); }}>
              Upgrade to Professional Plan
            </Button>
            <Dialog
              open={upgradeDialogue}
              fullWidth={true}
              maxWidth={"sm"}
            >
              <DialogTitle id="alert-dialog-title">{"Checkout"}</DialogTitle>
              <DialogContent>
                <UpdateBillingInfo
                organizationID={organizationID} route={"checkout"} closeDialogue={handleCloseDialogue} />
              </DialogContent>
              <DialogActions />
            </Dialog>
          </Grid>
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
      {data.billingInfo.billingPlan.billingPlanID === proPlan.billingPlanID && (
        <Grid container>
          <Grid item>
            <Button
              variant="outlined"
              className={classes.button}
              onClick={() => {setCancelDialogue(true); }}>
              Cancel plan
            </Button>
            <Dialog
              open={cancelDialogue}
              fullWidth={true}
              maxWidth={"sm"}
              onBackdropClick={() => { setCancelDialogue(false); }}
            >
              <DialogTitle id="alert-dialog-title">{"Are you sure?"}</DialogTitle>
              <DialogContent>
                <UpdateBillingInfo
                organizationID={organizationID} route={"cancel"} closeDialogue={handleCloseDialogue} />
              </DialogContent>
              <DialogActions />
            </Dialog>
          </Grid>
          <Grid item>
            <Button
              variant="contained"
              color="primary"
              className={classes.button}
              onClick={() => {
                window.location.href = "https://docs.google.com/forms/d/e/1FAIpQLSdsO3kcT3yk0Cgc4MzkPR_d16jZiYQd7L0M3ZxGwdOYycGhIg/viewform?usp=sf_link";
              }}>
              Discuss Enterprise Plan
            </Button>
          </Grid>
        </Grid>)}
    </React.Fragment>
  );
};

export default ViewBillingInfo;
