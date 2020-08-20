import { useMutation } from "@apollo/client";
import { Field, Form, Formik } from "formik";
import React, { FC } from "react";
import validator from "validator";

import { Container } from "@material-ui/core";

import { STAGE_PROJECT } from "../../apollo/queries/project";
import { ProjectByOrganizationAndName_projectByOrganizationAndName } from "../../apollo/types/ProjectByOrganizationAndName";
import { StageProject, StageProjectVariables } from "../../apollo/types/StageProject";
import { handleSubmitMutation, TextField as FormikTextField } from "../formik";
import SubmitControl from "../forms/SubmitControl";

interface EditProjectProps {
  project: ProjectByOrganizationAndName_projectByOrganizationAndName;
}

const EditProject: FC<EditProjectProps> = ({ project }) => {
  const [stageProject] = useMutation<StageProject, StageProjectVariables>(STAGE_PROJECT);

  const initialValues = {
    displayName: project.displayName || "",
    site: project.site || "",
    description: project.description || "",
    photoURL: project.photoURL || "",
  };

  return (
    <Container maxWidth={"sm"}>
      <Formik
        initialValues={initialValues}
        onSubmit={async (values, actions) =>
          handleSubmitMutation(
            values,
            actions,
            stageProject({
              variables: {
                organizationName: project.organization.name,
                projectName: project.name,
                displayName: values.displayName,
                site: values.site,
                description: values.description,
                photoURL: values.photoURL,
              },
            })
          )
        }
      >
        {({ isSubmitting, status }) => (
          <Form>
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
              name="site"
              validate={(val: string) => {
                if (val && !validator.isURL(val)) {
                  return "Site must be a valid URL";
                }
              }}
              component={FormikTextField}
              type="url"
              label="Project website"
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
            <Field name="photoURL" component={FormikTextField} label="Photo URL" />
            <SubmitControl
              label="Save changes"
              createdOn={project.createdOn}
              updatedOn={project.updatedOn}
              errorAlert={status}
              disabled={isSubmitting}
            />
          </Form>
        )}
      </Formik>
    </Container>
  );
};

export default EditProject;
