import { useMutation } from "@apollo/client";
import { NextPage } from "next";
import { useRouter } from "next/router";
import React from "react";

import { Button, Container, makeStyles, TextField, Theme, Typography } from "@material-ui/core";

import { UPDATE_ORGANIZATION } from "../../apollo/queries/organization";
import { REGISTER_USER_CONSENT } from "../../apollo/queries/user";
import { RegisterUserConsent, RegisterUserConsentVariables } from "../../apollo/types/RegisterUserConsent";
import { UpdateOrganization, UpdateOrganizationVariables } from "../../apollo/types/UpdateOrganization";
import { withApollo } from "../../apollo/withApollo";
import CheckboxField from "../../components/CheckboxField";
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
  const router = useRouter();
  const classes = useStyles();

  const me = useMe();
  if (!me?.personalUser) {
    return <></>;
  }

  if (me.personalUser.consentTerms) {
    if (typeof window !== "undefined") {
      router.push("/");
    }
  }

  const [values, setValues] = React.useState({
    name: me.name,
    consentTerms: me.personalUser.consentTerms,
    consentNewsletter: me.personalUser.consentNewsletter,
  });

  const [updateOrganization, { loading, error }] = useMutation<UpdateOrganization, UpdateOrganizationVariables>(
    UPDATE_ORGANIZATION,
    {
      onCompleted: (data) => {
        // stop if not succesful
        if (!data) {
          return;
        }

        // username updated successfully, so register consent
        if (me.personalUser) {
          registerUserConsent({
            variables: {
              userID: me.personalUser.userID,
              terms: values.consentTerms,
              newsletter: values.consentNewsletter,
            },
          });
        }
      },
    }
  );

  const [registerUserConsent, { loading: loadingConsent, error: errorConsent }] = useMutation<
    RegisterUserConsent,
    RegisterUserConsentVariables
  >(REGISTER_USER_CONSENT, {
    onCompleted: (data) => {
      if (data.registerUserConsent) {
        router.push("/");
      }
    },
  });

  const handleChange = (name: string) => (event: any) => {
    setValues({ ...values, [name]: event.target.value });
  };

  return (
    <Page title="Welcome" contentMarginTop="normal">
      <Container maxWidth="sm">
        <Typography component="h1" variant="h1" gutterBottom>
          Welcome to Beneath!
        </Typography>
        <Typography gutterBottom>We just need a few more details to get you set up</Typography>
        {/* Username form */}
        <form
          className={classes.form}
          onSubmit={(e) => {
            e.preventDefault();
            updateOrganization({
              variables: {
                organizationID: me.organizationID,
                name: toBackendName(values.name),
              },
            });
          }}
        >
          <TextField
            id="name"
            label={"Pick a username"}
            value={toURLName(values.name)}
            margin="normal"
            helperText={
              error
                ? isNameError(error)
                  ? "Username already taken"
                  : `An error occurred: ${JSON.stringify(error)}`
                : "Lowercase letters, numbers and dashes allowed (min. 3 characters)"
            }
            error={!validateName(values.name) || !!isNameError(error)}
            required
            fullWidth
            onChange={handleChange("name")}
          />
          <CheckboxField
            label="Email me updates about Beneath"
            checked={values.consentNewsletter}
            margin="normal"
            fullWidth
            onChange={(event, checked) => {
              setValues({ ...values, consentNewsletter: checked });
            }}
          />
          <CheckboxField
            label={
              <span>
                I agree to the{" "}
                <Link href="https://about.beneath.dev/policies/terms/">terms of service</Link> and{" "}
                <Link href="https://about.beneath.dev/policies/privacy/">privacy policy</Link>
              </span>
            }
            checked={values.consentTerms}
            margin="normal"
            fullWidth
            onChange={(event, checked) => {
              setValues({ ...values, consentTerms: checked });
            }}
          />
          <Button
            type="submit"
            variant="outlined"
            color="primary"
            fullWidth
            className={classes.submitConsentButton}
            disabled={loading || loadingConsent || !values.consentTerms}
          >
            All done
          </Button>
          {errorConsent && (
            <Typography variant="body1" color="error">
              {`An error occurred: ${JSON.stringify(errorConsent)}`}
            </Typography>
          )}
        </form>
      </Container>
    </Page>
  );
};

export default withApollo(WelcomePage);

const validateName = (val: string) => {
  return val && val.length >= 3 && val.length <= 40 && val.match(/^[_\-a-z][_\-a-z0-9]+$/);
};

const isNameError = (error?: Error) => {
  return error && error.message.match(/duplicate key value violates unique constraint/);
};
