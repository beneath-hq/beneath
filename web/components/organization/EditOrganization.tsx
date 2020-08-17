import { useMutation, useQuery } from "@apollo/client";
import { useRouter } from "next/router";
import React, { FC } from "react";
import Moment from "react-moment";

import { Button, makeStyles, TextField, Typography } from "@material-ui/core";

import { UPDATE_ORGANIZATION } from "../../apollo/queries/organization";
import { REGISTER_USER_CONSENT } from "../../apollo/queries/user";
import { OrganizationByName_organizationByName_PrivateOrganization } from "../../apollo/types/OrganizationByName";
import { RegisterUserConsent, RegisterUserConsentVariables } from "../../apollo/types/RegisterUserConsent";
import { UpdateOrganization, UpdateOrganizationVariables } from "../../apollo/types/UpdateOrganization";
import { toBackendName, toURLName } from "../../lib/names";
import CheckboxField from "../CheckboxField";
import VSpace from "../VSpace";

const useStyles = makeStyles((theme) => ({
  submitButton: {
    marginTop: theme.spacing(1.5),
  },
}));

export interface EditOrganizationProps {
  organization: OrganizationByName_organizationByName_PrivateOrganization;
}

const EditOrganization: FC<EditOrganizationProps> = ({ organization }) => {
  const [values, setValues] = React.useState({
    name: organization.name || "",
    displayName: organization.displayName || "",
    description: organization.description || "",
    photoURL: organization.photoURL || "",
    consentNewsletter: organization.personalUser?.consentNewsletter,
  });

  const router = useRouter();

  const [updateOrganization, { loading, error }] = useMutation<UpdateOrganization, UpdateOrganizationVariables>(
    UPDATE_ORGANIZATION, {
    onCompleted: (data) => {
      if (data.updateOrganization) {
        const name = toURLName(data.updateOrganization.name);
        if (name !== router.query.organization_name) {
          const href = `/organization?organization_name=${name}&tab=edit`;
          const as = `/${name}/-/edit`;
          router.replace(href, as, { shallow: true });
        }
      }
    },
  });

  const [registerUserConsent, { loading: loadingConsent, error: errorConsent }] = useMutation<
    RegisterUserConsent,
    RegisterUserConsentVariables
  >(REGISTER_USER_CONSENT);

  const handleChange = (name: string) => (event: any) => {
    setValues({ ...values, [name]: event.target.value });
  };

  const classes = useStyles();
  return (
    <div>
      <form
        onSubmit={(e) => {
          e.preventDefault();
          updateOrganization({
            variables: {
              organizationID: organization.organizationID,
              name: toBackendName(values.name),
              displayName: values.displayName,
              description: values.description,
              photoURL: values.photoURL,
            },
          });
          if (organization.personalUser?.userID) {
            if (values.consentNewsletter !== organization.personalUser.consentNewsletter) {
              registerUserConsent({
                variables: {
                  userID: organization.personalUser.userID,
                  newsletter: values.consentNewsletter,
                },
              });
            }
          }
        }}
      >
        <TextField
          id="name"
          label={organization.personalUser ? "Username" : "Name"}
          value={toURLName(values.name)}
          margin="normal"
          helperText={
            !validateName(values.name)
              ? "Lowercase letters, numbers and underscores allowed (minimum 3 characters)"
              : undefined
          }
          error={!validateName(values.name)}
          fullWidth
          required
          onChange={handleChange("name")}
        />
        <TextField
          id="display-name"
          label="Display Name"
          value={values.displayName}
          margin="normal"
          helperText={!validateDisplayName(values.displayName) ? "Should be between 1 and 50 characters" : undefined}
          error={!validateDisplayName(values.displayName)}
          fullWidth
          required
          onChange={handleChange("displayName")}
        />
        <TextField
          id="description"
          label="Bio"
          value={values.description}
          margin="normal"
          helperText={!validateBio(values.description) ? "Should be less than 255 characters" : undefined}
          error={!validateBio(values.description)}
          fullWidth
          multiline
          rows={1}
          rowsMax={3}
          onChange={handleChange("description")}
        />
        <TextField
          id="photo-url"
          label="Photo URL"
          value={values.photoURL || ""}
          margin="normal"
          fullWidth
          disabled
          onChange={handleChange("photoURL")}
        />
        {organization.personalUser && (
          <TextField
            id="email"
            label="Email"
            value={organization.personalUser.email}
            margin="normal"
            fullWidth
            disabled
          />
        )}
        {organization.personalUser && (
          <CheckboxField
            label="Subscribe to newsletter"
            checked={values.consentNewsletter}
            margin="normal"
            fullWidth
            onChange={(event, checked) => {
              setValues({ ...values, consentNewsletter: checked });
            }}
          />
        )}
        <Button
          type="submit"
          variant="outlined"
          color="primary"
          className={classes.submitButton}
          disabled={
            loading ||
            loadingConsent ||
            !validateName(values.name) ||
            !validateDisplayName(values.displayName) ||
            !validateBio(values.description)
          }
        >
          Save changes
        </Button>
        {error && (
          <Typography variant="body1" color="error">
            {isNameError(error)
              ? organization.personalUser
                ? "Username already taken"
                : "Name already taken"
              : `An error occurred: ${JSON.stringify(error)}`}
          </Typography>
        )}
        {errorConsent && (
          <Typography variant="body1" color="error">
            {`An error occurred: ${JSON.stringify(errorConsent)}`}
          </Typography>
        )}
      </form>
      <VSpace units={2} />
      <Typography variant="subtitle1" color="textSecondary">
        Profile created <Moment fromNow date={organization.createdOn} /> and last updated{" "}
        <Moment fromNow date={organization.updatedOn} />.
      </Typography>
    </div>
  );
};

export default EditOrganization;

const validateName = (val: string) => {
  return val && val.length >= 3 && val.length <= 40 && val.match(/^[_\-a-z][_\-a-z0-9]+$/);
};

const validateDisplayName = (val: string) => {
  return val && val.length >= 1 && val.length <= 50;
};

const validateBio = (val: string) => {
  return val.length < 256;
};

const isNameError = (error: Error) => {
  return error && error.message.match(/duplicate key value violates unique constraint/);
};
