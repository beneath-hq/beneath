import { useMutation } from "@apollo/client";
import { Field, Form, Formik } from "formik";
import { useRouter } from "next/router";
import React, { FC } from "react";
import validator from "validator";

import { STAGE_PROJECT } from "../../apollo/queries/project";
import { StageProject, StageProjectVariables } from "../../apollo/types/StageProject";
import { toURLName, toBackendName } from "../../lib/names";
import { handleSubmitMutation, SelectField as FormikSelectField, TextField as FormikTextField } from "../formik";
import SubmitControl from "../forms/SubmitControl";
import { Typography } from "@material-ui/core";
import useMe from "../../hooks/useMe";

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
  const [stageProject] = useMutation<StageProject, StageProjectVariables>(STAGE_PROJECT, {
    onCompleted: (data) => {
      if (data?.stageProject) {
        const orgName = toURLName(data.stageProject.organization.name);
        const projName = toURLName(data.stageProject.name);
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
  };

  return (
    <Formik
      initialValues={initialValues}
      onSubmit={async (values, actions) =>
        handleSubmitMutation(
          values,
          actions,
          stageProject({
            variables: {
              organizationName: values.organization ? toBackendName(values.organization.name) : "",
              projectName: toBackendName(values.name),
              displayName: values.displayName,
              description: values.description,
              photoURL: values.photoURL,
            },
          })
        )
      }
    >
      {({ isSubmitting, status }) => (
        <Form>
          <Typography component="h2" variant="h1" gutterBottom>
            Create project
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
            getOptionLabel={(option: Organization) => option.name}
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
          <SubmitControl label="Save changes" errorAlert={status} disabled={isSubmitting} />
        </Form>
      )}
    </Formik>
  );
};

export default CreateProject;