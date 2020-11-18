import React, {FC} from "react";
import { Button, Dialog, DialogActions, DialogContent, DialogTitle, Step, StepLabel, Stepper, Typography } from "@material-ui/core";

import { OrganizationByName_organizationByName_PrivateOrganization } from "apollo/types/OrganizationByName";
import VSpace from "components/VSpace";
import SelectBillingPlan from "./billing-plan/SelectBillingPlan";
import { BillingInfo_billingInfo, BillingInfo_billingInfo_billingPlan } from "ee/apollo/types/BillingInfo";
import Checkout from "./Checkout";
import ViewTaxInfo from "./tax-info/ViewTaxInfo";
import ViewBillingMethods from "./billing-method/ViewBillingMethods";

interface EditBillingProps {
  organization: OrganizationByName_organizationByName_PrivateOrganization;
  billingInfo: BillingInfo_billingInfo;
  changePlanDialog: boolean;
  setChangePlanDialog: (value: boolean) => void;
  addCard: (value: boolean) => void;
  editTaxInfo: (value: boolean) => void;
}

const EditBilling: FC<EditBillingProps> = ({organization, billingInfo, changePlanDialog, setChangePlanDialog, addCard, editTaxInfo }) => {
  const [activeStep, setActiveStep] = React.useState(0);
  const [selectedBillingPlan, setSelectedBillingPlan] = React.useState<BillingInfo_billingInfo_billingPlan | null>(null);

  const steps = ['Select a plan', 'Choose your billing method', 'Provide tax information', 'Checkout'];
  const handleNext = () => {
    setActiveStep((prevActiveStep) => prevActiveStep + 1);
  };

  const handleBack = () => {
    setActiveStep((prevActiveStep) => prevActiveStep - 1);
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
        onBackdropClick={() => closeAndReset()}
      >
        <DialogTitle>
          Change your billing plan
          <VSpace units={3} />
          <Stepper activeStep={activeStep}>
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
              <SelectBillingPlan selectBillingPlan={setSelectedBillingPlan} selectedBillingPlan={selectedBillingPlan} billingInfo={billingInfo} />
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
              <ViewBillingMethods organization={organization} billingInfo={billingInfo} addCard={addCard} />
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
              <ViewTaxInfo organization={organization} editable editTaxInfo={editTaxInfo} />
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
              <Checkout organization={organization} billingMethod={billingInfo.billingMethod} selectedBillingPlan={selectedBillingPlan} handleBack={handleBack} closeAndReset={closeAndReset} />
            </>
          )}
        </DialogContent>
      </Dialog>
    </>
  );
};

export default EditBilling;

{/* <DialogContent>
  <CancelPlan
    organization={organization}
    closeDialogue={(confirmationMessage) => handleCloseDialogue(confirmationMessage)}
    billingInfo={billingInfo}
  />
</DialogContent> */}
