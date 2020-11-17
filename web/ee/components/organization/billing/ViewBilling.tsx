import _ from "lodash";
import { Dialog, DialogContent, DialogTitle, Grid, Link, Typography } from "@material-ui/core";
import { Alert } from "@material-ui/lab";
import { makeStyles } from "@material-ui/core/styles";
import dynamic from "next/dynamic";
import React, { FC } from "react";

import VSpace from "components/VSpace";
import { useQuery } from "@apollo/client";
import { BillingInfo, BillingInfoVariables } from "ee/apollo/types/BillingInfo";
import { QUERY_BILLING_INFO } from "ee/apollo/queries/billingInfo";
import { OrganizationByName_organizationByName_PrivateOrganization } from "apollo/types/OrganizationByName";
import ViewBillingPlan from "./view/ViewBillingPlan";
import ChangeBillingPlan from "./EditBilling";
import CancelBillingPlan from "./edit/CancelBillingPlan";
import ViewBillingMethods from "./view/ViewBillingMethods";
import ViewTaxInfo from "./view/ViewTaxInfo";
import EditTaxInfo from "./edit/EditTaxInfo";
import clsx from "clsx";

const useStyles = makeStyles((theme) => ({
  sectionTitle: {
    marginTop: theme.spacing(6),
    marginBottom: theme.spacing(3)
  },
  firstSectionTitle: {
    marginTop: theme.spacing(4),
  }
}));

export interface ViewBillingProps {
  organization: OrganizationByName_organizationByName_PrivateOrganization;
}

const ViewBilling: FC<ViewBillingProps> = ({ organization }) => {
  const classes = useStyles();
  const [changePlanDialog, setChangePlanDialog] = React.useState(false);
  const [cancelPlanDialog, setCancelPlanDialog] = React.useState(false);
  const [addCardDialog, setAddCardDialog] = React.useState(false);
  const [editTaxInfoDialog, setEditTaxInfoDialog] = React.useState(false);
  const DynamicCardForm = dynamic(() => import("./edit/CardForm"));

  const { loading, error, data } = useQuery<BillingInfo, BillingInfoVariables>(QUERY_BILLING_INFO, {
    context: { ee: true },
    variables: {
      organizationID: organization.organizationID,
    },
  });

  if (error) {
    return <p>Error: {JSON.stringify(error)}</p>;
  }

  if (!data) {
    return <></>;
  }

  // alert when you're viewing billing for your personal organization but your main billing is handled by another org
  const specialCase =
    organization.personalUser && organization.personalUser.billingOrganizationID !== organization.organizationID;

  return (
    <React.Fragment>
      {specialCase && (
        <>
          <Alert severity="info">
            Note that you are a member of the {organization.personalUser?.billingOrganization.displayName} organization,
            and so {organization.personalUser?.billingOrganization.displayName} is billed for your Beneath usage.
            However, you are individually billed for the activity of any services that you have not transferred to{" "}
            {organization.personalUser?.billingOrganization.displayName}.
            <br />
            <br /> If you have no active services managed by your user, then make sure to cancel any active billing plan
            on this page. Then you won't incur any unneccessary charges.
            <br />
            <br />
            If you have active services that you would like {
              organization.personalUser?.billingOrganization.displayName
            }{" "}
            to pay for, you need to transfer those services from your personal user to the{" "}
            {organization.personalUser?.billingOrganization.displayName} organization. See the docs{" "}
            <Link href="https://about.beneath.dev/docs/core-resources/services">here</Link>.
            <br />
            <br /> If you have active services that you would like to pay for yourself, and not assign to your
            organization, then you must ensure the billing information on this page covers you. If the services' usage
            exceeds the quotas of the Free tier, then you should upgrade your personal billing plan on this page.
          </Alert>
          <VSpace units={2} />
        </>
      )}
      {/* {confirmationMessage && (
        <Alert severity="info">{confirmationMessage}</Alert>
      )} */}

      <Alert severity="info">
        You can find detailed information about our billing plans{" "}
        <Link href="https://about.beneath.dev/enterprise">here</Link>.
      </Alert>
      <Typography variant="h2" className={clsx(classes.sectionTitle, classes.firstSectionTitle)}>
        Billing plan
      </Typography>
      <ViewBillingPlan organization={organization} cancelPlan={setCancelPlanDialog} changePlan={setChangePlanDialog} />
      <ChangeBillingPlan
        organization={organization}
        billingInfo={data.billingInfo}
        changePlanDialog={changePlanDialog}
        setChangePlanDialog={setChangePlanDialog}
        addCard={setAddCardDialog}
        editTaxInfo={setEditTaxInfoDialog}
      />
      <CancelBillingPlan organization={organization} openDialog={cancelPlanDialog} openDialogFn={setCancelPlanDialog} />

      <Grid container>
        <Grid item xs={12} md={6}>
          <Typography variant="h2" className={classes.sectionTitle}>
            Billing methods
          </Typography>
          {/* <Typography variant="body1" gutterBottom>
            You are signed up to pay with the active billing method at the end of each billing cycle. Cards will be charged on the day of, and wire payments will be expected within 15 days.
          </Typography> */}
          <ViewBillingMethods organization={organization} billingInfo={data.billingInfo} addCard={setAddCardDialog}/>
          <Dialog
            open={addCardDialog}
            fullWidth={true}
            maxWidth={"md"}
            onBackdropClick={() => {
              setAddCardDialog(false);
            }}
          >
            <DialogTitle id="alert-dialog-title">{"Add a credit card"}</DialogTitle>
            <DialogContent>
              <DynamicCardForm organization={organization} openDialogFn={setAddCardDialog} />
            </DialogContent>
          </Dialog>
        </Grid>
      </Grid>

      <Grid container>
        <Grid item xs={12} md={6}>
          <Typography variant="h2" className={classes.sectionTitle}>
            Tax info
          </Typography>
          {/* <Typography gutterBottom>
            For paid plans, this information is necessary to compute tax for customers in certain countries.
          </Typography> */}
          <ViewTaxInfo organization={organization} editable editTaxInfo={setEditTaxInfoDialog}/>
          <Dialog open={editTaxInfoDialog} onBackdropClick={() => setEditTaxInfoDialog(false)} fullWidth maxWidth="sm">
            <DialogTitle>Edit tax info</DialogTitle>
            <DialogContent>
              <EditTaxInfo organization={organization} billingInfo={data.billingInfo} editTaxInfo={setEditTaxInfoDialog}/>
            </DialogContent>
          </Dialog>
        </Grid>
      </Grid>
    </React.Fragment>
  );
};

export default ViewBilling;
