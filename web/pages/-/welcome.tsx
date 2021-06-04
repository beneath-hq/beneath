import { useMutation } from "@apollo/client";
import { makeStyles, Theme, Typography } from "@material-ui/core";
import { Field, Form, Formik } from "formik";
import { NextPage } from "next";
import { useRouter } from "next/router";
import React, { useEffect } from "react";

import { UPDATE_ORGANIZATION } from "../../apollo/queries/organization";
import { REGISTER_USER_CONSENT } from "../../apollo/queries/user";
import { RegisterUserConsent, RegisterUserConsentVariables } from "../../apollo/types/RegisterUserConsent";
import { UpdateOrganization, UpdateOrganizationVariables } from "../../apollo/types/UpdateOrganization";
import { withApollo } from "../../apollo/withApollo";
import SubmitControl from "../../components/forms/SubmitControl";
import {
  Checkbox as FormikCheckbox,
  handleSubmitMutation,
  TextField as FormikTextField,
} from "../../components/formik";
import { Link } from "../../components/Link";
import Page from "../../components/Page";
import useMe from "../../hooks/useMe";
import { toBackendName, toURLName } from "../../lib/names";

const useStyles = makeStyles((theme: Theme) => ({
  form: {
    marginTop: theme.spacing(2.5),
  },
  submitConsentButton: {
    marginTop: theme.spacing(5),
  },
}));

const WelcomePage: NextPage = () => {
  // TODO: The routing hacks to get this to redirect are a mess

  const router = useRouter();
  const classes = useStyles();

  const [updateOrganization] = useMutation<UpdateOrganization, UpdateOrganizationVariables>(UPDATE_ORGANIZATION);
  const [registerUserConsent] = useMutation<RegisterUserConsent, RegisterUserConsentVariables>(REGISTER_USER_CONSENT, {
    onCompleted: () => {
      router.replace("/");
    },
  });

  const me = useMe();
  if (!me?.personalUser) {
    return <></>;
  }
  const user = me?.personalUser;

  useEffect(() => {
    if (user.consentTerms) {
      if (typeof window !== "undefined") {
        router.replace("/");
      }
    }
  }, []);

  const initialValues = {
    name: toURLName(me.name),
    consentTerms: me.personalUser.consentTerms as boolean,
    consentNewsletter: me.personalUser.consentNewsletter,
  };

  return (
    <Page title="Welcome" contentMarginTop="normal" maxWidth="sm">
      <Typography component="h1" variant="h1" gutterBottom>
        Welcome to Beneath!
      </Typography>
      <Typography gutterBottom>We just need a few more details to get you set up</Typography>
      {/* Username form */}
      <Formik
        initialValues={initialValues}
        onSubmit={(values, actions) => {
          return handleSubmitMutation(
            values,
            actions,
            // first run update username, and only update consent if successful
            (async () => {
              const res = await updateOrganization({
                variables: {
                  organizationID: me.organizationID,
                  name: toBackendName(values.name),
                },
              });
              if (res.errors) {
                return res;
              }
              return await registerUserConsent({
                variables: {
                  userID: user.userID,
                  terms: values.consentTerms,
                  newsletter: values.consentNewsletter,
                },
              });
            })()
          );
        }}
      >
        {({ isSubmitting, status, values }) => (
          <Form className={classes.form}>
            <Field
              name="name"
              validate={(val: string) => {
                if (!val || val.length < 3 || val.length > 40) {
                  return "Usernames should be between 3 and 40 characters long";
                }
                if (!val.match(/^[_\-a-z][_\-a-z0-9]+$/)) {
                  return "Usernames should consist of lowercase letters, numbers, underscores and dashes (cannot start with a number)";
                }
              }}
              component={FormikTextField}
              label={"Username"}
              required
            />

            <Field
              name="consentNewsletter"
              component={FormikCheckbox}
              type="checkbox"
              label="Email me updates about Beneath"
            />
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
                <span>
                  I agree to the{" "}
                  <Link href="https://about.beneath.dev/policies/terms/" target="_blank">
                    terms of service
                  </Link>{" "}
                  and{" "}
                  <Link href="https://about.beneath.dev/policies/privacy/" target="_blank">
                    privacy policy
                  </Link>
                </span>
              }
            />
            <SubmitControl label="All done" errorAlert={status} disabled={!values.consentTerms || isSubmitting} />
          </Form>
        )}
      </Formik>
    </Page>
  );
};

export default withApollo(WelcomePage);
