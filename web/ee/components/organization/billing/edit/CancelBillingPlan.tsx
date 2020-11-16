import _ from "lodash";
import React, { FC } from "react";

import {
  Button,
  Dialog,
  DialogContent,
  DialogContentText,
  Grid,
  Typography,
} from "@material-ui/core";

import { useMutation, useQuery } from "@apollo/client";
import { UPDATE_BILLING_PLAN } from "ee/apollo/queries/billingInfo";
import { QUERY_BILLING_PLANS } from "ee/apollo/queries/billingPlan";
import { QUERY_ORGANIZATION } from "apollo/queries/organization";
import { BillingPlans } from "ee/apollo/types/BillingPlans";
import { OrganizationByName_organizationByName_PrivateOrganization } from "apollo/types/OrganizationByName";
import { UpdateBillingPlan, UpdateBillingPlanVariables } from "ee/apollo/types/UpdateBillingPlan";

interface Props {
  organization: OrganizationByName_organizationByName_PrivateOrganization;
  openDialog: boolean;
  closeDialog: (value: boolean) => void;
}

const CancelBillingPlan: FC<Props> = ({ organization, openDialog, closeDialog }) => {
  const { loading, error: queryError, data: data1 } = useQuery<BillingPlans>(QUERY_BILLING_PLANS, {
    context: { ee: true },
  });

  const [updateBillingPlan, { error: mutError }] = useMutation<UpdateBillingPlan, UpdateBillingPlanVariables>(
    UPDATE_BILLING_PLAN,
    {
      context: { ee: true },
      onCompleted: (data) => {
        if (data) {
          // closeDialog("Your plan has been canceled.");
          closeDialog(true);
        }
      },
      refetchQueries: [{ query: QUERY_ORGANIZATION, variables: { name: organization.name } }],
      awaitRefetchQueries: true,
    }
  );

  if (queryError || !data1) {
    return <p>Error: {JSON.stringify(queryError)}</p>;
  }

  const freePlan = data1.billingPlans.filter((billingPlan) => billingPlan.default)[0];

  return (
    <>
      {/* TODO: use Formik, and add a checkbox to confirm that you want to cancel and know that you'll be charged for overage */}
      <Dialog open={openDialog}>
        <DialogContentText id="alert-dialog-description">
          Upon canceling your plan, your usage will be assessed and you will be charged for any applicable overage
          fees for the current billing period.
        </DialogContentText>
        <Grid container spacing={2}>
          <Grid item>
            <Button color="primary" autoFocus onClick={() => closeDialog(true)}>
              No, go back
            </Button>
          </Grid>
          <Grid item>
            <Button
              color="primary"
              autoFocus
              onClick={() => {
                updateBillingPlan({
                  variables: {
                    organizationID: organization.organizationID,
                    billingPlanID: freePlan.billingPlanID,
                  },
                });
              }}
            >
              Yes, I'm sure
            </Button>
          </Grid>
        </Grid>
      </Dialog>
    </>
  );
};

export default CancelBillingPlan;
