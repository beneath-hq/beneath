import { useMutation } from "@apollo/client";
import { Grid, Typography } from "@material-ui/core";
import { FC } from "react";

import { QUERY_ORGANIZATION } from "apollo/queries/organization";
import { OrganizationByName_organizationByName_PrivateOrganization } from "apollo/types/OrganizationByName";
import { handleSubmitMutation } from "components/formik";
import FormikRadioGroup from "components/formik/RadioGroup";
import FormikSelectField from "components/formik/SelectField";
import FormikTextField from "components/formik/TextField";
import SubmitControl from "components/forms/SubmitControl";
import { UPDATE_BILLING_DETAILS } from "ee/apollo/queries/billingInfo";
import { BillingInfo_billingInfo } from "ee/apollo/types/BillingInfo";
import { UpdateBillingDetails, UpdateBillingDetailsVariables } from "ee/apollo/types/UpdateBillingDetails";
import { COUNTRIES, US_STATES } from "ee/lib/billing";
import { Field, Form, Formik } from "formik";

interface Props {
  organization: OrganizationByName_organizationByName_PrivateOrganization;
  billingInfo: BillingInfo_billingInfo;
  editTaxInfo: (value: boolean) => void;
}

const EditTaxInfo: FC<Props> = ({organization, billingInfo, editTaxInfo}) => {
  const [updateBillingDetails] = useMutation<UpdateBillingDetails, UpdateBillingDetailsVariables>(
    UPDATE_BILLING_DETAILS,
    {
      context: { ee: true },
      onCompleted: (data) => {
        if (data) {
          editTaxInfo(false);
        }
      },
      refetchQueries: [{ query: QUERY_ORGANIZATION, variables: { name: organization.name } }],
      awaitRefetchQueries: true,
    }
  );

  const initialValues = {
    country: billingInfo.country ? billingInfo.country : null,
    region: billingInfo.region ? billingInfo.region : null,
    taxEntity: billingInfo.taxNumber ? "company" : "individual",
    companyName: billingInfo.companyName,
    taxNumber: billingInfo.taxNumber,
  };

  const sanitize = (values: any) => {
    if (values.country !== "United States of America") {
      values.region = "";
    }
    if (values.taxEntity === "individual") {
      values.companyName = "";
      values.taxNumber = "";
    }
  };

  return (
    <Formik
      initialValues={initialValues}
      onSubmit={async (values, actions) => {
        sanitize(values);
        handleSubmitMutation(
          values,
          actions,
          updateBillingDetails({
            variables: {
              organizationID: organization.organizationID,
              country: values.country,
              region: values.region,
              companyName: values.companyName,
              taxNumber: values.taxNumber,
            }
          })
        );
      }}
    >
      {({ values, isSubmitting, status }) => (
        <Form title="Edit tax info">
          <Typography variant="body2">
            Ensure this information is accurate so we can correctly compute tax for customers in certain jurisdictions.
          </Typography>
          <Grid container justify="space-between" spacing={2}>
            <Grid item xs={6}>
              <Field
                name="country"
                component={FormikSelectField}
                label="Country"
                required
                options={COUNTRIES}
                getOptionLabel={(option: string) => option}
                getOptionSelected={(option: string, value: string) => {
                  return option === value;
                }}
              />
            </Grid>
            {values.country === "United States of America" && (
              <Grid item xs={6}>
                <Field
                  name="region"
                  component={FormikSelectField}
                  label="State"
                  required
                  options={US_STATES}
                  getOptionLabel={(option: string) => (option)}
                  getOptionSelected={(option: string, value: string) => {
                    return option === value;
                  }}
                />
              </Grid>
            )}
          </Grid>
          <Field
            name="taxEntity"
            component={FormikRadioGroup}
            label="Tax Entity"
            required
            options={[
              { value: "company", label: "Company" },
              { value: "individual", label: "Individual" },
            ]}
            row
          />
          {values.taxEntity === "company" && (
            <>
            <Grid container justify="space-between" spacing={2}>
              <Grid item xs={6}>
                <Field
                  name="companyName"
                  label="Company Name"
                  component={FormikTextField}
                  required
                  validate={(val: string) => {
                    if (!val) {
                      return "Please provide your company's name";
                    }
                    if (val && val.length > 50) {
                      return "Company names should be shorter than 50 characters";
                    }
                  }}
                />
              </Grid>
              <Grid item xs={6}>
                <Field
                  name="taxNumber"
                  label="Tax ID"
                  component={FormikTextField}
                  required
                  validate={(val: string) => {
                    if (!val) {
                      return "Please provide your company's TaxID";
                    }
                    if (val && val.length > 50) {
                      return "TaxIDs should be shorter than 50 characters";
                    }
                  }}
                />
              </Grid>
            </Grid>
            </>
          )}
          <SubmitControl label="Done" errorAlert={status} disabled={isSubmitting} />
        </Form>
      )}
    </Formik>
  );
};

export default EditTaxInfo;