import { useMutation } from "@apollo/client";
import { makeStyles, Typography } from "@material-ui/core";
import { Alert, AlertTitle } from "@material-ui/lab";
import { Field, Formik } from "formik";
import React, { FC } from "react";

import { ISSUE_USER_SECRET, QUERY_USER_SECRETS } from "apollo/queries/secret";
import { IssueUserSecret, IssueUserSecretVariables } from "apollo/types/IssueUserSecret";
import { Form, handleSubmitMutation, RadioGroup as FormikRadioGroup, TextField as FormikTextField } from "components/formik";
import SubmitControl from "components/forms/SubmitControl";
import CodeBlock from "components/CodeBlock";

const useStyles = makeStyles((theme) => ({
  newSecretCard: {
    marginTop: theme.spacing(3),
  },
}));

export interface IssueSecretProps {
  userID: string;
}

const IssueSecret: FC<IssueSecretProps> = ({ userID }) => {
  const [newSecretString, setNewSecretString] = React.useState("");

  const [issueSecret, { loading, error }] = useMutation<IssueUserSecret, IssueUserSecretVariables>(ISSUE_USER_SECRET, {
    onCompleted: (data) => {
      setNewSecretString(data.issueUserSecret.token);
    },
  });

  const initialValues = {
    description: "",
    access: "full",
    readOnly: false,
    publicOnly: false,
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
            issueSecret({
              variables: {
                description: values.description,
                readOnly: values.access !== "full",
                publicOnly: values.access === "readpublic",
              },
              update: (cache, { data }) => {
                if (data) {
                  // reset form
                  actions.resetForm();

                  // update cache
                  const queryData = cache.readQuery({
                    query: QUERY_USER_SECRETS,
                    variables: { userID },
                  }) as any;
                  cache.writeQuery({
                    query: QUERY_USER_SECRETS,
                    variables: { userID },
                    data: { secretsForUser: [data.issueUserSecret.secret].concat(queryData.secretsForUser) },
                  });
                }
              },
            })
          )
        }
      >
        {({ isSubmitting, status, values }) => (
          <Form
            title="Create personal secret"
            variant="embedded"
            helperText={
              <>
                Personal secrets have all your access permissions and their usage is counted directly against your
                personal quotas. You should only use them in your local environment. Use a{" "}
                <strong>service secret</strong> when deploying to production or embedding secrets in public code.
              </>
            }
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
            <Field
              name="access"
              component={FormikRadioGroup}
              label="Access"
              required
              options={[
                { value: "full", label: "Full" },
                { value: "readonly", label: "Private read" },
                { value: "readpublic", label: "Public read" },
              ]}
              row
            />
            {values.access === "full" && (
              <Typography variant="body2" color="textSecondary">
                Full access secrets can read/write streams and edit resources. Use them for e.g. CLI authentication,
                private scripts and Jupyter notebooks.
                <br />
                <strong>Keep secure and don't share with others.</strong>
              </Typography>
            )}
            {values.access === "readonly" && (
              <Typography variant="body2" color="textSecondary">
                Private read secrets can read data from every public and private stream you have access to.
              </Typography>
            )}
            {values.access === "readpublic" && (
              <Typography variant="body2" color="textSecondary">
                Public read secrets can read data from every public stream you have access to, but not from streams in
                private projects.
              </Typography>
            )}
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
