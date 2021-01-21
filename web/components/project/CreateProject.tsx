import { useMutation } from "@apollo/client";
import { Typography } from "@material-ui/core";
import { Field, Formik } from "formik";
import { useRouter } from "next/router";
import React, { FC } from "react";

import { CREATE_PROJECT } from "apollo/queries/project";
import {
  CreateProject as ApolloCreateProject,
  CreateProjectVariables as ApolloCreateProjectVariables,
} from "apollo/types/CreateProject";
import {
  Form,
  handleSubmitMutation,
  RadioGroup as FormikRadioGroup,
  SelectField as FormikSelectField,
  TextField as FormikTextField,
} from "components/formik";
import SubmitControl from "components/forms/SubmitControl";
import useMe from "hooks/useMe";
import { toURLName, toBackendName } from "lib/names";

interface Organization {
  organizationID: string;
  name: string;
  displayName?: string;
}

export interface CreateProjectProps {
  preselectedOrganization?: Organization;
}

const CreateProject: FC<CreateProjectProps> = ({ preselectedOrganization }) => {
  const router = useRouter();
  const [createProject] = useMutation<ApolloCreateProject, ApolloCreateProjectVariables>(CREATE_PROJECT, {
    onCompleted: (data) => {
      if (data?.createProject) {
        const orgName = toURLName(data.createProject.organization.name);
        const projName = toURLName(data.createProject.name);
        const href = `/project?organization_name=${orgName}&project_name=${projName}`;
        const as = `/${orgName}/${projName}`;
        router.replace(href, as, { shallow: true });
      }
    },
  });

  // create the list of organizations you can pick from
  const me = useMe();
  const organizations: Organization[] = [];
  if (me) {
    organizations.push(me);
    const billingOrg = me.personalUser?.billingOrganization;
    if (billingOrg && billingOrg.organizationID !== me.organizationID) {
      organizations.push(billingOrg);
    }
  }

  const initialValues = {
    organization:
      organizations.length === 0 ? null : preselectedOrganization ? preselectedOrganization : organizations[0],
    name: "",
    displayName: "",
    description: "",
    photoURL: "",
    public: "private",
  };

  return (
    <Formik
      initialValues={initialValues}
      onSubmit={async (values, actions) =>
        handleSubmitMutation(
          values,
          actions,
          createProject({
            variables: {
              input: {
                organizationID: values?.organization?.organizationID || "",
                projectName: toBackendName(values.name),
                displayName: values.displayName,
                description: values.description,
                photoURL: values.photoURL,
                public: values.public === "public" ? true : false,
              },
            },
          })
        )
      }
    >
      {({ values, isSubmitting, status }) => (
        <Form title="Create project">
          <Typography variant="body2">
            Projects contain streams and services, and every project has its own access management. A project in Beneath
            is like a repo in Git.
          </Typography>
          <Field
            name="organization"
            validate={(org?: Organization) => {
              if (!org || !org?.name) {
                return "Projects must have an owner";
              }
            }}
            component={FormikSelectField}
            label="Owner"
            helperText="Select a user or organization"
            required
            options={organizations}
            getOptionLabel={(option: Organization) => toURLName(option.name)}
            getOptionSelected={(option: Organization, value: Organization) => {
              return option.name === value.name;
            }}
          />
          <Field
            name="name"
            validate={(val: string) => {
              if (!val || val.length < 3 || val.length > 40) {
                return "Project names should be between 3 and 40 characters long";
              }
              if (!val.match(/^[_\-a-z][_\-a-z0-9]+$/)) {
                return "Project names should consist of lowercase letters, numbers, underscores and dashes (cannot start with a number)";
              }
            }}
            component={FormikTextField}
            helperText={`Your project URL will be https://beneath.dev/${
              values.organization ? toURLName(values.organization.name) : "USERNAME"
            }/${values.name ? toURLName(values.name) : "NAME"}`}
            label="Name"
            required
          />
          <Field
            name="displayName"
            validate={(val: string) => {
              if (val && val.length > 50) {
                return "Display names should be shorter than 50 characters";
              }
            }}
            component={FormikTextField}
            label="Display Name"
          />
          <Field
            name="description"
            validate={(val: string) => {
              if (val && val.length > 255) {
                return "Bios should be shorter than 255 characters";
              }
            }}
            component={FormikTextField}
            label="Description"
            multiline
            rows={1}
            rowsMax={3}
          />
          <Field
            name="public"
            component={FormikRadioGroup}
            label="Access"
            required
            options={[
              { value: "public", label: "Public" },
              { value: "private", label: "Private" },
            ]}
            row
          />
          {values.public === "public" && (
            <Typography variant="body2" color="textSecondary">
              Open your data streams to the world and see what people build!
            </Typography>
          )}
          {values.public === "private" && (
            <Typography variant="body2" color="textSecondary">
              Keep your data streams private and add collaborators as needed.
            </Typography>
          )}
          <SubmitControl label="Create project" errorAlert={status} disabled={isSubmitting} />
        </Form>
      )}
    </Formik>
  );
};

export default CreateProject;
