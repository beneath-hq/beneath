import { useMutation } from "@apollo/client";
import { Container, Typography } from "@material-ui/core";
import { Field, Form, Formik } from "formik";
import React, { FC } from "react";
import validator from "validator";

import { UPDATE_PROJECT } from "apollo/queries/project";
import { ProjectByOrganizationAndName_projectByOrganizationAndName } from "apollo/types/ProjectByOrganizationAndName";
import { UpdateProject, UpdateProjectVariables } from "apollo/types/UpdateProject";
import { handleSubmitMutation, RadioGroup as FormikRadioGroup, TextField as FormikTextField } from "components/formik";
import SubmitControl from "components/forms/SubmitControl";

interface EditProjectProps {
  project: ProjectByOrganizationAndName_projectByOrganizationAndName;
}

const EditProject: FC<EditProjectProps> = ({ project }) => {
  const [updateProject] = useMutation<UpdateProject, UpdateProjectVariables>(UPDATE_PROJECT);

  const initialValues = {
    displayName: project.displayName || "",
    site: project.site || "",
    description: project.description || "",
    photoURL: project.photoURL || "",
    public: project.public ? "public" : "private",
  };

  return (
    <Container maxWidth={"sm"}>
      <Formik
        initialValues={initialValues}
        onSubmit={async (values, actions) =>
          handleSubmitMutation(
            values,
            actions,
            updateProject({
              variables: {
                input: {
                  projectID: project.projectID,
                  displayName: values.displayName,
                  site: values.site,
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
          <Form>
            {/* <Field
              name="displayName"
              validate={(val: string) => {
                if (val && val.length > 50) {
                  return "Display names should be shorter than 50 characters";
                }
              }}
              component={FormikTextField}
              label="Display Name"
            /> */}
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
                Open your tables to the world and see what people build!
              </Typography>
            )}
            {values.public === "private" && (
              <Typography variant="body2" color="textSecondary">
                Keep your tables private and add collaborators as needed.
              </Typography>
            )}
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
