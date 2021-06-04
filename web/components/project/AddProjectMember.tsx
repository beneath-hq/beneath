import { useApolloClient, useMutation } from "@apollo/client";
import { Grid, makeStyles, Typography } from "@material-ui/core";
import { Field, Formik } from "formik";
import React, { FC, useState } from "react";

import {
  UpdateUserProjectPermissions,
  UpdateUserProjectPermissionsVariables,
} from "apollo/types/UpdateUserProjectPermissions";
import {
  OrganizationByName,
  OrganizationByNameVariables,
  OrganizationByName_organizationByName,
} from "apollo/types/OrganizationByName";
import { QUERY_ORGANIZATION } from "apollo/queries/organization";
import { QUERY_PROJECT_MEMBERS, UPDATE_USER_PROJECT_PERMISSIONS } from "apollo/queries/project";
import BetterAvatar from "components/Avatar";
import { Form, handleSubmitMutation, TextField as FormikTextField } from "components/formik";
import FormikRadioGroup from "components/formik/RadioGroup";
import SubmitControl from "components/forms/SubmitControl";
import { toURLName } from "lib/names";

const useStyles = makeStyles((theme) => ({
  profileContainer: {
    marginTop: theme.spacing(2),
    marginBottom: theme.spacing(1),
  },
}));

export interface FindUserProps {
  onChange: (user?: OrganizationByName_organizationByName) => void;
}

export const FindUser: FC<FindUserProps> = ({ onChange }) => {
  const apollo = useApolloClient();
  const initialValues = { name: "" };

  return (
    <Formik
      initialValues={initialValues}
      onSubmit={async (values, actions) => {
        const { data, error, errors } = await apollo.query<OrganizationByName, OrganizationByNameVariables>({
          query: QUERY_ORGANIZATION,
          variables: values,
        });

        if (data?.organizationByName.personalUserID) {
          onChange(data.organizationByName);
        } else {
          if (error || errors) {
            actions.setFieldError("name", "Couldn't find user");
          } else if (data) {
            actions.setFieldError("name", "That's an organization, not a user");
          }
          onChange(undefined);
        }

        actions.setSubmitting(false);
      }}
    >
      {({ isSubmitting }) => (
        <Form title="Add member" variant="embedded">
          <Field
            name="name"
            label="Username"
            required
            component={FormikTextField}
            validate={(val: string) => {
              if (!val || val.length < 3 || val.length > 40) {
                return "User names are between 3 and 40 characters long";
              }
              if (!val.match(/^[_\-a-z][_\-a-z0-9]+$/)) {
                return "User names consist of lowercase letters, numbers, underscores and dashes (cannot start with a number)";
              }
            }}
          />
          <SubmitControl label="Find user" color="secondary" disabled={isSubmitting} />
        </Form>
      )}
    </Formik>
  );
};

export interface AddProjectMemberProps {
  projectID: string;
  onCompleted?: () => void;
}

const AddProjectMember: FC<AddProjectMemberProps> = ({ projectID, onCompleted }) => {
  const [user, setUser] = useState<OrganizationByName_organizationByName | undefined>(undefined);

  const [updateUserProjectPermissions] = useMutation<
    UpdateUserProjectPermissions,
    UpdateUserProjectPermissionsVariables
  >(UPDATE_USER_PROJECT_PERMISSIONS, {
    onError: (error) => {
      console.log(error);
    },
    onCompleted: (data) => {
      if (data.updateUserProjectPermissions) {
        if (onCompleted) {
          onCompleted();
        }
      }
    },
  });

  const initialValues = {
    view: "false",
    create: "false",
    admin: "false",
  };

  const classes = useStyles();
  return (
    <>
      <FindUser onChange={setUser} />
      {user && (
        <Grid container spacing={2} alignItems="center" className={classes.profileContainer}>
          <Grid item>
            <BetterAvatar size="list" label={user.name} src={user.photoURL} />
          </Grid>
          <Grid item>
            <Typography>@{toURLName(user.name)}</Typography>
          </Grid>
          <Grid item>
            <Typography>{user.displayName}</Typography>
          </Grid>
        </Grid>
      )}
      {user && (
        <Formik
          initialValues={initialValues}
          onSubmit={(values, actions) =>
            handleSubmitMutation(
              values,
              actions,
              updateUserProjectPermissions({
                variables: {
                  projectID,
                  userID: user?.personalUserID || "",
                  view: values.view === "true" ? true : false,
                  create: values.create === "true" ? true : false,
                  admin: values.admin === "true" ? true : false,
                },
                refetchQueries: [
                  {
                    query: QUERY_PROJECT_MEMBERS,
                    variables: { projectID },
                  },
                ],
              })
            )
          }
        >
          {({ isSubmitting, status }) => (
            <Form variant="embedded">
              <Field
                name="view"
                component={FormikRadioGroup}
                label="View access"
                required
                options={[
                  { value: "true", label: "True" },
                  { value: "false", label: "False" },
                ]}
                row
              />
              <Field
                name="create"
                component={FormikRadioGroup}
                label="Create access"
                required
                options={[
                  { value: "true", label: "True" },
                  { value: "false", label: "False" },
                ]}
                row
              />
              <Field
                name="admin"
                component={FormikRadioGroup}
                label="Admin access"
                required
                options={[
                  { value: "true", label: "True" },
                  { value: "false", label: "False" },
                ]}
                row
              />
              <SubmitControl label="Add member" errorAlert={status} disabled={isSubmitting} />
            </Form>
          )}
        </Formik>
      )}
    </>
  );
};

export default AddProjectMember;
