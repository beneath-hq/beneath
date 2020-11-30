import { useQuery } from "@apollo/client";
import dynamic from "next/dynamic";
import React, {FC} from "react";
import { Box, Button, Dialog, DialogActions, DialogContent, DialogTitle, Grid, makeStyles, Step, StepLabel, Stepper, Typography } from "@material-ui/core";

import VSpace from "components/VSpace";
import { BillingInfo, BillingInfoVariables, BillingInfo_billingInfo_billingPlan } from "ee/apollo/types/BillingInfo";
import { QUERY_BILLING_INFO } from "ee/apollo/queries/billingInfo";
import SelectBillingPlan from "./billing-plan/SelectBillingPlan";
import ViewBillingMethods from "./billing-method/ViewBillingMethods";
import Finalize from "./Finalize";
import ViewTaxInfo from "./tax-info/ViewTaxInfo";
import CancelBillingPlan from "./billing-plan/CancelBillingPlan";
import EditTaxInfo from "./tax-info/EditTaxInfo";
import { OrganizationByName_organizationByName_PrivateOrganization } from "apollo/types/OrganizationByName";

const useStyles = makeStyles((theme) => ({
  stepper: {
    backgroundColor: theme.palette.background.default,
    overflowX: "auto",
    paddingLeft: theme.spacing(0),
    paddingRight: theme.spacing(0),
  },
  box: {
    height: theme.spacing(70)
  }
}));

interface Props {
  organization: OrganizationByName_organizationByName_PrivateOrganization;
}

const Checkout: FC<Props> = ({organization}) => {
  const classes = useStyles();
  const [activeStep, setActiveStep] = React.useState(0);
  const [selectedBillingPlan, setSelectedBillingPlan] = React.useState<BillingInfo_billingInfo_billingPlan | null>(null);
  const [addCardDialog, setAddCardDialog] = React.useState(false);
  const [editTaxInfoDialog, setEditTaxInfoDialog] = React.useState(false);
  const DynamicCardForm = dynamic(() => import("./billing-method/CardForm"));

  const { loading, error, data } = useQuery<BillingInfo, BillingInfoVariables>(QUERY_BILLING_INFO, {
    context: { ee: true },
    variables: {
      organizationID: organization.organizationID,
    },
  });

  if (!data) return null;
  const billingInfo = data.billingInfo;

  const steps = ['Select a plan', 'Choose your billing method', 'Provide tax information', 'Finalize'];
  const handleNext = () => {
    // if the Free plan is selected, skip the steps for billing method & tax info
    if (selectedBillingPlan && selectedBillingPlan.default && activeStep === 0) {
      setActiveStep(3);
      return;
    }
    setActiveStep(activeStep + 1);
  };

  const handleBack = () => {
    // if the Free plan is selected, skip the steps for billing method & tax info
    if (selectedBillingPlan && selectedBillingPlan.default && activeStep === 3) {
      setActiveStep(0);
      return;
    }
    setActiveStep(activeStep - 1);
  };

  return (
    <>
      <Typography variant="h1">
        Change your billing plan
      </Typography>
      <VSpace units={3} />
      <Stepper activeStep={activeStep} className={classes.stepper}>
        {steps.map((step) => (
          <Step key={step}>
            <StepLabel>
              {step}
            </StepLabel>
          </Step>
        ))}
      </Stepper>
      <VSpace units={6} />
      <Box className={classes.box}>
        {activeStep === 0 && (
          <>
            <SelectBillingPlan selectBillingPlan={setSelectedBillingPlan} selectedBillingPlan={selectedBillingPlan} billingInfo={billingInfo} />
          </>
        )}
        {activeStep === 1 && (
          <>
            <Grid container justify="center">
              <Grid item xs={8}>
                <ViewBillingMethods organization={organization} billingInfo={billingInfo} addCard={setAddCardDialog} />
              </Grid>
              <Dialog
                open={addCardDialog}
                fullWidth={true}
                maxWidth={"sm"}
                onBackdropClick={() => setAddCardDialog(false)}
              >
                <DialogTitle id="alert-dialog-title">{"Add a card"}</DialogTitle>
                <DialogContent>
                  <DynamicCardForm organization={organization} openDialogFn={setAddCardDialog} />
                </DialogContent>
              </Dialog>
            </Grid>
          </>
        )}
        {activeStep === 2 && (
          <>
            <Grid container justify="center">
              <ViewTaxInfo organization={organization} editable editTaxInfo={setEditTaxInfoDialog} />
              <Dialog open={editTaxInfoDialog} onBackdropClick={() => setEditTaxInfoDialog(false)} fullWidth maxWidth="sm">
                <DialogTitle>Edit tax info</DialogTitle>
                <DialogContent>
                  <EditTaxInfo organization={organization} billingInfo={data.billingInfo} editTaxInfo={setEditTaxInfoDialog}/>
                </DialogContent>
              </Dialog>
            </Grid>
          </>
        )}
        {activeStep === 3 && selectedBillingPlan && billingInfo.billingMethod && (
          <>
            {!selectedBillingPlan.default && (
              <Finalize organization={organization} billingMethod={billingInfo.billingMethod} selectedBillingPlan={selectedBillingPlan} handleBack={handleBack} />
            )}
            {selectedBillingPlan.default && (
              <CancelBillingPlan organization={organization} handleBack={handleBack} />
            )}
          </>
        )}
      </Box>

      {/* Buttons */}
      <Grid container justify="flex-end" spacing={2}>
        {activeStep === 0 && (
          <>
            <Grid item>
              <Button onClick={handleNext} disabled={!selectedBillingPlan || selectedBillingPlan.billingPlanID === billingInfo.billingPlan.billingPlanID}>
                Next
              </Button>
            </Grid>
          </>
        )}
        {activeStep === 1 && (
          <>
            <Grid item>
              <Button onClick={handleBack}>
                Back
              </Button>
            </Grid>
            <Grid item>
              <Button onClick={handleNext} disabled={!billingInfo.billingMethod}>
                Next
              </Button>
            </Grid>
          </>
        )}
        {activeStep === 2 && (
          <>
            <Grid item>
              <Button onClick={handleBack}>
                Back
              </Button>
            </Grid>
            <Grid item>
              <Button onClick={handleNext} disabled={!billingInfo.country}>
                Next
              </Button>
            </Grid>
          </>
        )}
        {activeStep === 3 && (
          <>
          </>
        )}
      </Grid>
    </>
  );
};

export default Checkout;