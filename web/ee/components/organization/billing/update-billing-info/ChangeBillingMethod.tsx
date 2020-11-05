import _ from "lodash";
import React, { FC } from "react";

import { Typography } from "@material-ui/core";

import { useMutation, useQuery } from "@apollo/client";
import { UPDATE_BILLING_METHOD } from "ee/apollo/queries/billingInfo";
import { QUERY_BILLING_METHODS } from "ee/apollo/queries/billingMethod";
import { BillingInfo_billingInfo } from "ee/apollo/types/BillingInfo";
import { BillingMethods, BillingMethodsVariables } from "ee/apollo/types/BillingMethods";
import { UpdateBillingMethod, UpdateBillingMethodVariables } from "ee/apollo/types/UpdateBillingMethod";
import billing from "ee/lib/billing";
import { Formik, Form, Field } from "formik";
import { handleSubmitMutation } from "components/formik";
import SubmitControl from "components/forms/SubmitControl";
import FormikSelectField from "components/formik/SelectField";

interface BillingMethod {
  billingMethodID: string;
  paymentsDriver: string;
  driverPayload: string;
}

interface Props {
  closeDialogue: (confirmationMessage: string) => void;
  billingInfo: BillingInfo_billingInfo;
}

const ChangeBillingMethod: FC<Props> = ({ closeDialogue, billingInfo }) => {
  const { error, data } = useQuery<BillingMethods, BillingMethodsVariables>(QUERY_BILLING_METHODS, {
    context: { ee: true },
    variables: { organizationID: billingInfo.organizationID },
  });

  const [updateBillingMethod] = useMutation<UpdateBillingMethod, UpdateBillingMethodVariables>(UPDATE_BILLING_METHOD, {
    context: { ee: true },
    onCompleted: (data) => {
      if (data) {
        closeDialogue("Your billing method has been changed.");
      }
    },
  });

  if (error || !data) {
    return <p>Error: {JSON.stringify(error)}</p>;
  }

  const billingMethods: BillingMethod[] = data.billingMethods;

  const computeLabel = (billingMethod: BillingMethod) => {
    if (billingMethod.paymentsDriver === billing.STRIPECARD_DRIVER) {
      const payload = JSON.parse(billingMethod.driverPayload);
      return payload.brand.charAt(0).toUpperCase() + payload.brand.slice(1) + " xxxx-xxxx-xxxx-" + payload.last4;
    } else if (billingMethod.paymentsDriver === billing.STRIPEWIRE_DRIVER) {
      return "Wire payment";
    } else if (billingMethod.paymentsDriver === billing.ANARCHISM_DRIVER) {
      return "Anarchy!";
    } else {
      return "Unrecognized payment driver.";
    }
  };

  const initialValues: { billingMethod: BillingMethod | null } = { billingMethod: null };

  return (
    <Formik
      initialValues={initialValues}
      onSubmit={async (values, actions) =>
        handleSubmitMutation(
          values,
          actions,
          updateBillingMethod({
            variables: {
              organizationID: billingInfo.organizationID,
              billingMethodID: values.billingMethod?.billingMethodID,
            },
          })
        )
      }
    >
      {({ isSubmitting, status }) => (
        <Form>
          <Typography component="h2" variant="h1" gutterBottom>
            Change billing method
          </Typography>
          <Field
            name="billingMethod"
            validate={(billingMethod?: BillingMethod) => {
              if (!billingMethod) {
                return "Please set a billing method.";
              }
            }}
            component={FormikSelectField}
            label="Billing method"
            helperText="Select one of your billing methods on file"
            required
            options={billingMethods}
            getOptionLabel={computeLabel}
            getOptionSelected={(option: BillingMethod, value: BillingMethod) => {
              return option.billingMethodID === value.billingMethodID;
            }}
          />
          <SubmitControl label="Save changes" errorAlert={status} disabled={isSubmitting} />
        </Form>
      )}
    </Formik>
  );
};

export default ChangeBillingMethod;
