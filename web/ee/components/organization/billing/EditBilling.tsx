import React, {FC} from "react";
import { Button, Dialog, DialogActions, DialogContent, DialogTitle, Step, StepLabel, Stepper, Typography } from "@material-ui/core";

import { OrganizationByName_organizationByName_PrivateOrganization } from "apollo/types/OrganizationByName";
import VSpace from "components/VSpace";
import SelectBillingPlan from "./edit/SelectBillingPlan";
import { BillingInfo_billingInfo, BillingInfo_billingInfo_billingPlan } from "ee/apollo/types/BillingInfo";
import Checkout from "./edit/Checkout";
import ViewBillingMethods from "./view/ViewBillingMethods";
import ViewTaxInfo from "./view/ViewTaxInfo";

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

  return (
    <>
    {/* should this be in a Dialog, or could it just replace the content on View Billing? or is that wrong to not explicitly change the url? */}
      <Dialog open={changePlanDialog} fullWidth={true} maxWidth={"md"} onBackdropClick={() => setChangePlanDialog(false)}>
        <DialogTitle>
          Change your billing plan
          <Typography>Choose the plan that best suits your needs.</Typography>
        </DialogTitle>
        <DialogContent dividers={false}>
          <Stepper activeStep={activeStep}>
            {steps.map((step) => (
              <Step key={step}>
                <StepLabel>
                  {step}
                </StepLabel>
              </Step>
            ))}
          </Stepper>
          {activeStep === 0 && (
            <>
              <SelectBillingPlan selectBillingPlan={setSelectedBillingPlan} billingInfo={billingInfo} />
              <DialogActions>
                <Button onClick={() => setChangePlanDialog(false)}>
                  Cancel
                </Button>
                <Button onClick={handleNext} disabled={!selectedBillingPlan}>
                  Next
                </Button>
              </DialogActions>
            </>
          )}
          {activeStep === 1 && (
            <>
              <VSpace units={4} />
              <Typography gutterBottom>
                Select an active billing method
              </Typography>
              <ViewBillingMethods organization={organization} billingInfo={billingInfo} addCard={addCard} />
              <VSpace units={4} />
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
              <VSpace units={4} />
              <ViewTaxInfo organization={organization} editable editTaxInfo={editTaxInfo} />
              <VSpace units={4} />
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
              <VSpace units={4} />
              <Checkout organization={organization} billingMethod={billingInfo.billingMethod} selectedBillingPlan={selectedBillingPlan} handleBack={handleBack} setChangePlanDialog={setChangePlanDialog} />
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
