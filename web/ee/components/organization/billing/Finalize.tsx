import _ from "lodash";
import React, { FC } from "react";
import { Grid, Link, makeStyles, Theme } from "@material-ui/core";

import { OrganizationByName_organizationByName_PrivateOrganization } from "apollo/types/OrganizationByName";
import ViewTaxInfo from "./tax-info/ViewTaxInfo";
import ViewBillingMethod from "./billing-method/ViewBillingMethod";
import {
  BillingInfo_billingInfo_billingMethod,
  BillingInfo_billingInfo_billingPlan,
} from "ee/apollo/types/BillingInfo";
import FormikCheckbox from "components/formik/Checkbox";
import SubmitControl from "components/forms/SubmitControl";
import { Field, Form, Formik } from "formik";
import { handleSubmitMutation } from "components/formik";
import { UpdateBillingPlan, UpdateBillingPlanVariables } from "ee/apollo/types/UpdateBillingPlan";
import { useMutation } from "@apollo/client";
import { QUERY_BILLING_INFO, UPDATE_BILLING_PLAN } from "ee/apollo/queries/billingInfo";
import { QUERY_ORGANIZATION } from "apollo/queries/organization";
import ViewBillingPlanDescription from "./billing-plan/ViewBillingPlanDescription";
import VSpace from "components/VSpace";
import { useRouter } from "next/router";
import { toURLName } from "lib/names";

const useStyles = makeStyles((theme: Theme) => ({
  sectionTitle: {
    marginBottom: theme.spacing(2),
  },
}));

interface Props {
  organization: OrganizationByName_organizationByName_PrivateOrganization;
  billingMethod: BillingInfo_billingInfo_billingMethod;
  selectedBillingPlan: BillingInfo_billingInfo_billingPlan;
  handleBack: () => void;
}

const Finalize: FC<Props> = ({ organization, billingMethod, selectedBillingPlan, handleBack }) => {
  const classes = useStyles();
  const router = useRouter();
  const [updateBillingPlan] = useMutation<UpdateBillingPlan, UpdateBillingPlanVariables>(UPDATE_BILLING_PLAN, {
    context: { ee: true },
    onCompleted: (data) => {
      if (data.updateBillingPlan) {
        const orgName = toURLName(organization.name);
        const href = `/organization?organization_name=${orgName}&tab=billing`;
        const as = `/${orgName}/-/billing`;
        router.replace(href, as);
      }
    },
    refetchQueries: [
      { query: QUERY_ORGANIZATION, variables: { name: organization.name } },
      { query: QUERY_BILLING_INFO, variables: { organizationID: organization.organizationID }, context: { ee: true } },
    ],
    // awaitRefetchQueries: true, // This causes a React error: "Can't perform a React state update on an unmounted component"
  });

  const initialValues = {
    consentTerms: false,
  };

  return (
    <>
      <Grid container spacing={2}>
        <Grid item xs={12} md={4}>
          <ViewBillingPlanDescription billingPlan={selectedBillingPlan} />
        </Grid>
        <Grid item xs={12} md={4}>
          <ViewBillingMethod
            paymentsDriver={billingMethod.paymentsDriver}
            driverPayload={billingMethod.driverPayload}
          />
        </Grid>
        <Grid item xs={12} md={4}>
          <ViewTaxInfo organization={organization} littleHeader />
        </Grid>
      </Grid>
      <VSpace units={3} />
      <Formik
        initialValues={initialValues}
        onSubmit={(values, actions) =>
          handleSubmitMutation(
            values,
            actions,
            updateBillingPlan({
              variables: {
                organizationID: organization.organizationID,
                billingPlanID: selectedBillingPlan.billingPlanID,
              },
            })
          )
        }
      >
        {({ isSubmitting, status, values }) => (
          <Form>
            <Field
              name="consentTerms"
              component={FormikCheckbox}
              type="checkbox"
              validate={(checked: any) => {
                if (!checked) {
                  return "Cannot continue without consent to the terms of service";
                }
              }}
              label={
                // the Stripe docs recommend this text when you're collecting a card to charge on a recurring basis
                <span>
                  I authorise Beneath to send instructions to the financial institution that issued my card to take
                  payments from my card account in accordance with the
                  <Link href="https://about.beneath.dev/policies/"> terms </Link> of my agreement with you.
                </span>
              }
            />
            <SubmitControl
              label="Purchase"
              cancelFn={handleBack}
              cancelLabel="Back"
              rightSide
              errorAlert={status}
              disabled={!values.consentTerms || isSubmitting}
            />
          </Form>
        )}
      </Formik>
    </>
  );
};

export default Finalize;
