import { useMutation, useQuery } from "@apollo/client";
import _ from "lodash";
import React, { FC } from "react";
import { Field, Formik } from "formik";
import {
  Dialog,
  DialogContent,
} from "@material-ui/core";

import { OrganizationByName_organizationByName_PrivateOrganization } from "apollo/types/OrganizationByName";
import { handleSubmitMutation, Form } from "components/formik";
import FormikCheckbox from "components/formik/Checkbox";
import SubmitControl from "components/forms/SubmitControl";
import { QUERY_BILLING_INFO, UPDATE_BILLING_PLAN } from "ee/apollo/queries/billingInfo";
import { QUERY_BILLING_PLANS } from "ee/apollo/queries/billingPlan";
import { QUERY_ORGANIZATION } from "apollo/queries/organization";
import { BillingPlans } from "ee/apollo/types/BillingPlans";
import { UpdateBillingPlan, UpdateBillingPlanVariables } from "ee/apollo/types/UpdateBillingPlan";

interface Props {
  organization: OrganizationByName_organizationByName_PrivateOrganization;
  openDialog: boolean;
  openDialogFn: (value: boolean) => void;
}

const CancelBillingPlan: FC<Props> = ({ organization, openDialog, openDialogFn }) => {
  const { loading, error, data } = useQuery<BillingPlans>(QUERY_BILLING_PLANS, {
    context: { ee: true },
  });

  const [updateBillingPlan] = useMutation<UpdateBillingPlan, UpdateBillingPlanVariables>(
    UPDATE_BILLING_PLAN,
    {
      context: { ee: true },
      onCompleted: (data) => {
        if (data) {
          // openDialogFn("Your plan has been canceled.");
          openDialogFn(false); // close the dialog
        }
      },
      refetchQueries: [
        { query: QUERY_ORGANIZATION, variables: { name: organization.name } },
        { query: QUERY_BILLING_INFO, variables: { organizationID: organization.organizationID }, context: { ee: true } },
      ],
      awaitRefetchQueries: true,
    }
  );

  if (error) {
    return <p>Error: {JSON.stringify(error)}</p>;
  }

  if (!data) {
    return <></>;
  }

  const freePlan = data.billingPlans.filter((billingPlan) => billingPlan.default)[0];

  const initialValues = {
    confirmCancel: false,
  };

  return (
    <>
      <Dialog open={openDialog}>
        <DialogContent>
          <Formik
            initialValues={initialValues}
            onSubmit={(values, actions) =>
              handleSubmitMutation(
                values,
                actions,
                updateBillingPlan({
                  variables: {
                    organizationID: organization.organizationID,
                    billingPlanID: freePlan.billingPlanID
                  }
                })
              )
            }
          >
            {({ isSubmitting, status, values }) => (
              <Form title="Cancel plan">
                <Field
                  name="confirmCancel"
                  component={FormikCheckbox}
                  type="checkbox"
                  validate={(checked: any) => {
                    if (!checked) {
                      return "Cannot continue without confirming you want to cancel";
                    }
                  }}
                  label={
                    <span>
                      I understand that my current usage will be assessed, and, if applicable, I'll be billed for any overages in the current period.
                    </span>
                  }
                />
                <SubmitControl label="Cancel plan" severe cancelFn={() => openDialogFn(false)} cancelLabel="Back" errorAlert={status} disabled={!values.confirmCancel || isSubmitting} />
              </Form>
            )}
          </Formik>
        </DialogContent>
      </Dialog>
    </>
  );
};

export default CancelBillingPlan;
