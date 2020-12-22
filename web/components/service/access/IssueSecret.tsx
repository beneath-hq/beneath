import { useMutation } from "@apollo/client";
import { Field, Formik } from "formik";
import React, { FC } from "react";

import { ISSUE_SERVICE_SECRET, QUERY_SERVICE_SECRETS } from "apollo/queries/secret";
import { IssueServiceSecret, IssueServiceSecretVariables } from "apollo/types/IssueServiceSecret";
import { Form, handleSubmitMutation, TextField as FormikTextField } from "components/formik";
import SubmitControl from "components/forms/SubmitControl";

export interface Props {
  serviceID: string;
  onCompleted: (newSecretString: string) => void;
}

const IssueSecret: FC<Props> = ({ serviceID, onCompleted }) => {
  const [issueServiceSecret, { loading, error }] = useMutation<IssueServiceSecret, IssueServiceSecretVariables>(ISSUE_SERVICE_SECRET, {
    onCompleted: (data) => {
      onCompleted(data.issueServiceSecret.token);
    },
  });

  const initialValues = {
    description: "",
  };

  return (
    <>
      <Formik
        initialValues={initialValues}
        onSubmit={(values, actions) =>
          handleSubmitMutation(
            values,
            actions,
            issueServiceSecret({
              variables: {
                serviceID: serviceID,
                description: values.description,
              },
              update: (cache, { data }) => {
                if (data) {
                  // update cache
                  const queryData = cache.readQuery({
                    query: QUERY_SERVICE_SECRETS,
                    variables: { serviceID },
                  }) as any;
                  cache.writeQuery({
                    query: QUERY_SERVICE_SECRETS,
                    variables: { serviceID },
                    data: { secretsForService: [data.issueServiceSecret.secret].concat(queryData.secretsForService) },
                  });
                }
              },
            })
          )
        }
      >
        {({ isSubmitting, status, values }) => (
          <Form
            title="Create service secret"
            variant="embedded"
          >
            <Field
              name="description"
              validate={(description: string) => {
                if (description.length === 0) {
                  return "You must provide a description";
                }
                if (description.length > 40) {
                  return "Descriptions should be less than 40 characters long";
                }
              }}
              component={FormikTextField}
              label="Description"
              placeholder="My service secret"
              required
            />
            <SubmitControl label="Create secret" errorAlert={status} disabled={isSubmitting} />
          </Form>
        )}
      </Formik>
    </>
  );
};

export default IssueSecret;
