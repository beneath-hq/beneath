import { useMutation } from "@apollo/client";
import { makeStyles, Typography } from "@material-ui/core";
import { Alert, AlertTitle } from "@material-ui/lab";
import { Field, Formik } from "formik";
import React, { FC } from "react";

import { ISSUE_SERVICE_SECRET, QUERY_SERVICE_SECRETS } from "apollo/queries/secret";
import { IssueServiceSecret, IssueServiceSecretVariables } from "apollo/types/IssueServiceSecret";
import { Form, handleSubmitMutation, TextField as FormikTextField } from "components/formik";
import SubmitControl from "components/forms/SubmitControl";

const useStyles = makeStyles((theme) => ({
  newSecretCard: {
    marginTop: theme.spacing(3),
  },
}));

export interface Props {
  serviceID: string;
}

const IssueSecret: FC<Props> = ({ serviceID }) => {
  const [newSecretString, setNewSecretString] = React.useState("");

  const [issueServiceSecret, { loading, error }] = useMutation<IssueServiceSecret, IssueServiceSecretVariables>(ISSUE_SERVICE_SECRET, {
    onCompleted: (data) => {
      setNewSecretString(data.issueServiceSecret.token);
    },
  });

  const initialValues = {
    description: "",
  };

  const classes = useStyles();
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
                  // reset form
                  actions.resetForm();

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
              placeholder="My personal secret"
              required
            />
            <SubmitControl label="Create secret" errorAlert={status} disabled={isSubmitting} />
          </Form>
        )}
      </Formik>
      {newSecretString !== "" && (
        <Alert severity="success" className={classes.newSecretCard}>
          <AlertTitle>Here is your new secret!</AlertTitle>
          <Typography gutterBottom variant="body2">
            {newSecretString}
          </Typography>
          <Typography>
            The secret will only be shown this once – remember to keep it safe!
          </Typography>
        </Alert>
      )}
    </>
  );
};

export default IssueSecret;
