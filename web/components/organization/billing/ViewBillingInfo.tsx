import { useQuery } from "@apollo/client";
import _ from "lodash";
import React, { FC } from "react";

import { QUERY_BILLING_INFO } from "../../../apollo/queries/billinginfo";
import { BillingInfo, BillingInfo_billingInfo, BillingInfoVariables } from "../../../apollo/types/BillingInfo";
import { OrganizationByName_organizationByName_PrivateOrganization } from "../../../apollo/types/OrganizationByName";

import { Button, Dialog, DialogActions, DialogContent, DialogTitle, Grid, Paper, Typography } from "@material-ui/core";
import { makeStyles } from "@material-ui/core/styles";

import billing from "../../../lib/billing";
import Loading from "../../Loading";
import CancelPlan from "./update-billing-info/CancelPlan";
import ChangeBillingMethod from "./update-billing-info/ChangeBillingMethod";
import Checkout from "./update-billing-info/Checkout";

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
  banner: {
    padding: theme.spacing(2),
    marginBottom: theme.spacing(2),
  },
}));

export interface BillingInfoProps {
  organization: OrganizationByName_organizationByName_PrivateOrganization;
}

const ViewBillingInfo: FC<BillingInfoProps> = ({ organization }) => {
  const classes = useStyles();
  const [upgradeDialogue, setUpgradeDialogue] = React.useState(false);
  const [changeBillingMethodDialogue, setChangeBillingMethodDialogue] = React.useState(false);
  const [cancelDialogue, setCancelDialogue] = React.useState(false);
  const [confirmationMessage, setConfirmationMessage] = React.useState("");

  const { loading, error, data } = useQuery<BillingInfo, BillingInfoVariables>(QUERY_BILLING_INFO, {
    variables: {
      organizationID: organization.organizationID,
    },
  });

  if (error || !data) {
    return <p>Error: {JSON.stringify(error)}</p>;
  }

  const billingInfo = data.billingInfo;

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

  const displayPrepaidQuota = (prepaidQuota: any, isBillingPlanDefault: boolean) => {
    if (isBillingPlanDefault) {
      return "N/A";
    } else if (!prepaidQuota) {
      return "none";
    } else {
      return (prepaidQuota / 10 ** 9).toString() + " GB";
    }
  };

  const billingPlan = [
    { name: "Plan name", detail: billingInfo.billingPlan.description },
    {
      name: "Prepaid read quota",
      detail: displayPrepaidQuota(organization.prepaidReadQuota, billingInfo.billingPlan.default),
    },
    {
      name: "Prepaid write quota",
      detail: displayPrepaidQuota(organization.prepaidWriteQuota, billingInfo.billingPlan.default),
    },
    {
      name: "Prepaid scan quota",
      detail: displayPrepaidQuota(organization.prepaidScanQuota, billingInfo.billingPlan.default),
    },
    {
      name: "Read quota",
      detail: organization.readQuota ? (organization.readQuota / 10 ** 9).toString() + " GB" : "unlimited",
    },
    {
      name: "Write quota",
      detail: organization.writeQuota ? (organization.writeQuota / 10 ** 9).toString() + " GB" : "unlimited",
    },
    {
      name: "Scan quota",
      detail: organization.scanQuota ? (organization.scanQuota / 10 ** 9).toString() + " GB" : "unlimited",
    },
  ];

  const taxInfo = [
    { name: "Country", detail: billingInfo.country },
    { name: "Region", detail: billingInfo.region },
    { name: "Company", detail: billingInfo.companyName },
    { name: "Tax ID", detail: billingInfo.taxNumber },
  ];

  const handleCloseDialogue = (confirmationMessage: string) => {
    setConfirmationMessage(confirmationMessage);
    setUpgradeDialogue(false);
    setChangeBillingMethodDialogue(false);
    setCancelDialogue(false);
  };

  return (
    <>
      {confirmationMessage && (
        <Paper elevation={1} square>
          <Typography className={classes.banner}>{confirmationMessage}</Typography>
        </Paper>
      )}
      <Grid container spacing={3}>
        <Grid item xs={12} md={4}>
          <Grid container>
            <Grid item>
              <Typography variant="h5" className={classes.title}>
                Billing plan
              </Typography>
            </Grid>
          </Grid>
          {billingPlan.map((billingPlan) => (
            <React.Fragment key={billingPlan.name}>
              <Grid container>
                <Grid item xs={6}>
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
          <Grid container alignItems="center">
            <Grid item>
              <Typography variant="h5" className={classes.title}>
                Billing method
              </Typography>
            </Grid>
            {!billingInfo.billingPlan.default && (
              <Grid item>
                <Button
                  color="primary"
                  onClick={() => {
                    setChangeBillingMethodDialogue(true);
                  }}
                  className={classes.title}
                >
                  Edit
                </Button>
                <Dialog
                  open={changeBillingMethodDialogue}
                  fullWidth={true}
                  maxWidth={"sm"}
                  onBackdropClick={() => {
                    setChangeBillingMethodDialogue(false);
                  }}
                >
                  <DialogTitle id="alert-dialog-title">{"Change billing method"}</DialogTitle>
                  <DialogContent>
                    <ChangeBillingMethod
                      organization={organization}
                      closeDialogue={(confirmationMessage) => handleCloseDialogue(confirmationMessage)}
                      billingInfo={billingInfo}
                    />
                  </DialogContent>
                  <DialogActions />
                </Dialog>
              </Grid>
            )}
          </Grid>
          <Typography>{displayBillingMethod(billingInfo)}</Typography>
        </Grid>
        <Grid item xs={12} md={4}>
          <Grid container>
            <Grid item>
              <Typography variant="h5" className={classes.title}>
                Tax info
              </Typography>
            </Grid>
          </Grid>
          {billingInfo.billingPlan.default && <Typography>N/A</Typography>}
          {!billingInfo.billingPlan.default &&
            taxInfo.map((taxInfo) => (
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
      {billingInfo.billingPlan.default && (
        <Grid container>
          <Grid item>
            <Button
              variant="outlined"
              color="primary"
              className={classes.button}
              onClick={() => { setUpgradeDialogue(true); }}
            >
              Upgrade to Professional Plan
            </Button>
            <Dialog open={upgradeDialogue} fullWidth={true} maxWidth={"md"}>
              <DialogTitle id="alert-dialog-title">{"Checkout"}</DialogTitle>
              <DialogContent>
                <Checkout
                  organization={organization}
                  closeDialogue={(confirmationMessage) => handleCloseDialogue(confirmationMessage)}
                />
              </DialogContent>
              <DialogActions />
            </Dialog>
          </Grid>
          <Grid item>
            <Button
              variant="outlined"
              className={classes.button}
              href="https://docs.google.com/forms/d/e/1FAIpQLSdsO3kcT3yk0Cgc4MzkPR_d16jZiYQd7L0M3ZxGwdOYycGhIg/viewform?usp=sf_link"
            >
              Discuss Enterprise Plan
            </Button>
          </Grid>
        </Grid>
      )}
      {!billingInfo.billingPlan.default && billingInfo.billingPlan.availableInUI && (
        <Grid container>
          <Grid item>
            <Button
              variant="outlined"
              className={classes.button}
              onClick={() => {
                setCancelDialogue(true);
              }}
            >
              Cancel plan
            </Button>
            <Dialog
              open={cancelDialogue}
              fullWidth={true}
              maxWidth={"sm"}
              onBackdropClick={() => {
                setCancelDialogue(false);
              }}
            >
              <DialogTitle id="alert-dialog-title">{"Are you sure you want to cancel?"}</DialogTitle>
              <DialogContent>
                <CancelPlan
                  organization={organization}
                  closeDialogue={(confirmationMessage) => handleCloseDialogue(confirmationMessage)}
                  billingInfo={billingInfo}
                />
              </DialogContent>
              <DialogActions />
            </Dialog>
          </Grid>
          <Grid item>
            <Button
              variant="outlined"
              color="primary"
              className={classes.button}
              onClick={() => {
                window.location.href =
                  "https://docs.google.com/forms/d/e/1FAIpQLSdsO3kcT3yk0Cgc4MzkPR_d16jZiYQd7L0M3ZxGwdOYycGhIg/viewform?usp=sf_link";
              }}
            >
              Discuss Enterprise Plan
            </Button>
          </Grid>
        </Grid>
      )}
    </>
  );
};

export default ViewBillingInfo;
