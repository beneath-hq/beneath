import { useMutation, useQuery } from "@apollo/client";
import _ from "lodash";
import React, { FC } from "react";
import { Field, Formik } from "formik";
import {
  Grid,
  Typography,
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
import { toURLName } from "lib/names";
import { useRouter } from "next/router";

interface Props {
  organization: OrganizationByName_organizationByName_PrivateOrganization;
  handleBack: () => void;
}

const CancelBillingPlan: FC<Props> = ({ organization, handleBack }) => {
  const router = useRouter();
  const { loading, error, data } = useQuery<BillingPlans>(QUERY_BILLING_PLANS, {
    context: { ee: true },
  });

  const [updateBillingPlan] = useMutation<UpdateBillingPlan, UpdateBillingPlanVariables>(
    UPDATE_BILLING_PLAN,
    {
      context: { ee: true },
      onCompleted: (data) => {
        if (data) {
          const orgName = toURLName(organization.name);
          const href = `/organization/-/billing?organization_name=${orgName}`;
          const as = `/${orgName}/-/billing`;
          router.replace(href, as, {shallow: true});
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
          <Form>
            <Grid container>
              <Grid item xs={3}>
              </Grid>
              <Grid item xs={6}>
                <Typography variant="h2">Cancel plan</Typography>
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
              </Grid>
            </Grid>
            <SubmitControl label="Cancel plan" severe rightSide cancelFn={handleBack} cancelLabel="Back" errorAlert={status} disabled={!values.confirmCancel || isSubmitting} />
          </Form>
        )}
      </Formik>
    </>
  );
};

export default CancelBillingPlan;
