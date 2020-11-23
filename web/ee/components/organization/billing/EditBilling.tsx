import React, {FC} from "react";
import { Button, Dialog, DialogActions, DialogContent, DialogTitle, Grid, makeStyles, Step, StepLabel, Stepper } from "@material-ui/core";

import { OrganizationByName_organizationByName_PrivateOrganization } from "apollo/types/OrganizationByName";
import VSpace from "components/VSpace";
import { BillingInfo_billingInfo, BillingInfo_billingInfo_billingPlan } from "ee/apollo/types/BillingInfo";
import SelectBillingPlan from "./billing-plan/SelectBillingPlan";
import ViewBillingMethods from "./billing-method/ViewBillingMethods";
import Checkout from "./Checkout";
import ViewTaxInfo from "./tax-info/ViewTaxInfo";
import CancelBillingPlan from "./billing-plan/CancelBillingPlan";

const useStyles = makeStyles((theme) => ({
  stepper: {
    backgroundColor: theme.palette.background.default,
    overflowX: "auto",
  }
}));

interface EditBillingProps {
  organization: OrganizationByName_organizationByName_PrivateOrganization;
  billingInfo: BillingInfo_billingInfo;
  changePlanDialog: boolean;
  setChangePlanDialog: (value: boolean) => void;
  addCard: (value: boolean) => void;
  editTaxInfo: (value: boolean) => void;
}

const EditBilling: FC<EditBillingProps> = ({organization, billingInfo, changePlanDialog, setChangePlanDialog, addCard, editTaxInfo }) => {
  const classes = useStyles();
  const [activeStep, setActiveStep] = React.useState(0);
  const [selectedBillingPlan, setSelectedBillingPlan] = React.useState<BillingInfo_billingInfo_billingPlan | null>(null);

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

  const closeAndReset = () => {
    setActiveStep(0);
    setSelectedBillingPlan(null);
    setChangePlanDialog(false);
  };

  return (
    <>
    {/* should this be in a Dialog, or could it just replace the content on View Billing? or is that wrong to not explicitly change the url? */}
      <Dialog
        open={changePlanDialog}
        fullWidth={true}
        maxWidth={"md"}
      >
        <DialogTitle>
          Change your billing plan
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
        </DialogTitle>
        <DialogContent dividers={false}>
          {activeStep === 0 && (
            <>
              <Grid container>
                <Grid item xs={2}>
                </Grid>
                <Grid item xs={8}>
                  <SelectBillingPlan selectBillingPlan={setSelectedBillingPlan} selectedBillingPlan={selectedBillingPlan} billingInfo={billingInfo} />
                </Grid>
              </Grid>
              <DialogActions>
                <Button onClick={() => closeAndReset()}>
                  Cancel
                </Button>
                <Button onClick={handleNext} disabled={!selectedBillingPlan || selectedBillingPlan.billingPlanID === billingInfo.billingPlan.billingPlanID}>
                  Next
                </Button>
              </DialogActions>
            </>
          )}
          {activeStep === 1 && (
            <>
              <Grid container>
                <Grid item xs={3}>
                </Grid>
                <Grid item xs={6}>
                  <ViewBillingMethods organization={organization} billingInfo={billingInfo} addCard={addCard} />
                </Grid>
              </Grid>
              <DialogActions>
                <Button onClick={handleBack}>
                  Back
                </Button>
                <Button onClick={handleNext} disabled={!billingInfo.billingMethod}>
                  Next
                </Button>
              </DialogActions>
            </>
          )}
          {activeStep === 2 && (
            <>
              <Grid container>
                <Grid item xs={3}>
                </Grid>
                <Grid item xs={6}>
                  <ViewTaxInfo organization={organization} editable editTaxInfo={editTaxInfo} />
                </Grid>
              </Grid>
              <DialogActions>
                <Button onClick={handleBack}>
                  Back
                </Button>
                <Button onClick={handleNext} disabled={!billingInfo.country}>
                  Next
                </Button>
              </DialogActions>
            </>
          )}
          {activeStep === 3 && selectedBillingPlan && billingInfo.billingMethod && (
            <>
              {!selectedBillingPlan.default && (
                <Checkout organization={organization} billingMethod={billingInfo.billingMethod} selectedBillingPlan={selectedBillingPlan} handleBack={handleBack} closeAndReset={closeAndReset} />
              )}
              {selectedBillingPlan.default && (
                <CancelBillingPlan organization={organization} handleBack={handleBack} closeAndReset={closeAndReset} />
              )}
            </>
          )}
        </DialogContent>
      </Dialog>
    </>
  );
};

export default EditBilling;