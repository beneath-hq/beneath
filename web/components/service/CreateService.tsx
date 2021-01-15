import { useMutation, useQuery } from "@apollo/client";
import { Typography } from "@material-ui/core";
import { Field, Formik } from "formik";
import { useRouter } from "next/router";
import React, { FC } from "react";

import { QUERY_PROJECTS_FOR_USER } from "apollo/queries/project";
import {
  CreateService as ApolloCreateService,
  CreateServiceVariables as ApolloCreateServiceVariables,
} from "apollo/types/CreateService";
import {
  Form,
  handleSubmitMutation,
  SelectField as FormikSelectField,
  TextField as FormikTextField,
} from "components/formik";
import SubmitControl from "components/forms/SubmitControl";
import useMe from "hooks/useMe";
import { toURLName, toBackendName } from "lib/names";
import { ProjectsForUser, ProjectsForUserVariables } from "apollo/types/ProjectsForUser";
import { CREATE_SERVICE } from "apollo/queries/service";

interface Project {
  organization: { name: string };
  name: string;
  displayName?: string;
}

interface Props {
  preselectedProject?: Project;
}

const CreateService: FC<Props> = ({ preselectedProject }) => {
  const me = useMe();
  const router = useRouter();
  const [createService] = useMutation<ApolloCreateService, ApolloCreateServiceVariables>(CREATE_SERVICE, {
    onCompleted: (data) => {
      if (data?.createService) {
        const orgName = toURLName(data.createService.project.organization.name);
        const projName = toURLName(data.createService.project.name);
        const serviceName = toURLName(data.createService.name);
        const href = `/service?organization_name=${orgName}&project_name=${projName}&service_name=${serviceName}`;
        const as = `/${orgName}/${projName}/service:${serviceName}`;
        router.replace(href, as, { shallow: true });
      }
    },
  });

  const { data, loading, error } = useQuery<ProjectsForUser, ProjectsForUserVariables>(QUERY_PROJECTS_FOR_USER, {
    variables: { userID: me?.personalUserID || "" },
    skip: !me,
  });

  const initialValues = {
    project: data?.projectsForUser?.length ? (preselectedProject ? preselectedProject : data.projectsForUser[0]) : null,
    name: "",
    description: "",
    sourceURL: "",
    readQuota: "",
    writeQuota: "",
    scanQuota: "",
  };

  return (
    <Formik
      initialValues={initialValues}
      onSubmit={async (values, actions) =>
        handleSubmitMutation(
          values,
          actions,
          createService({
            variables: {
              input: {
                organizationName: toBackendName(values.project?.organization.name || ""),
                projectName: toBackendName(values.project?.name || ""),
                serviceName: toBackendName(values.name),
                description: values.description,
                sourceURL: values.sourceURL,
                readQuota: values.readQuota ? Number(values.readQuota) * 10 ** 9 : null,
                writeQuota: values.writeQuota ? Number(values.writeQuota) * 10 ** 9 : null,
                scanQuota: values.scanQuota ? Number(values.scanQuota) * 10 ** 9 : null,
              },
            },
          })
        )
      }
    >
      {({ values, isSubmitting, status }) => (
        <Form title="Create service">
          <Typography variant="body2">
            A service in Beneath is a *non-user account* (also known as a "service account" in e.g. GCP). Services are
            used to isolate the activity of a production piece of code. A service has its own quotas and (minimally
            viable) permissions.
            <br />
            <br />
            After creating a service with this form, you can grant it permissions and issue a secret to use in
            production.
          </Typography>
          <Field
            name="project"
            validate={(proj?: Project) => {
              if (!proj) {
                return "Select a project for the stream";
              }
            }}
            component={FormikSelectField}
            label="Project"
            required
            loading={loading}
            options={data?.projectsForUser || []}
            getOptionLabel={(option: Project) => `${toURLName(option.organization.name)}/${toURLName(option.name)}`}
            getOptionSelected={(option: Project, value: Project) => {
              return option.name === value.name;
            }}
          />
          <Field
            name="name"
            validate={(val: string) => {
              if (!val || val.length < 3 || val.length > 40) {
                return "Service names should be between 3 and 40 characters long";
              }
              if (!val.match(/^[_\-a-z][_\-a-z0-9]+$/)) {
                return "Service names should consist of lowercase letters, numbers, underscores and dashes (cannot start with a number)";
              }
            }}
            component={FormikTextField}
            label="Service name"
            required
            helperText={`URL preview: https://beneath.dev/${
              values.project?.organization.name ? toURLName(values.project.organization.name) : ""
            }/${values.project?.name ? toURLName(values.project.name) : ""}/service:${
              values.name ? toURLName(values.name) : ""
            }`}
          />
          <Field
            name="description"
            validate={(val: string) => {
              if (val && val.length > 255) {
                return "Descriptions should be shorter than 255 characters";
              }
            }}
            component={FormikTextField}
            label="Description"
            multiline
            rows={1}
            rowsMax={3}
          />
          {/* <Field
            name="sourceURL"
            validate={(val: string) => {
              // TODO: need to validate that it's a url
            }}
            component={FormikTextField}
            label="URL to source code"
            multiline
            rows={1}
            rowsMax={3}
          /> */}
          <Field
            name="readQuota"
            component={FormikTextField}
            label="Read Quota (GB)"
            validate={(val: string) => {
              if (!val.match(/^[0-9]*$/)) {
                return "Numbers only";
              }
            }}
          />
          <Field
            name="writeQuota"
            component={FormikTextField}
            label="Write Quota (GB)"
            validate={(val: string) => {
              if (!val.match(/^[0-9]*$/)) {
                return "Numbers only";
              }
            }}
          />
          <Field
            name="scanQuota"
            component={FormikTextField}
            label="Scan Quota (GB)"
            validate={(val: string) => {
              if (!val.match(/^[0-9]*$/)) {
                return "Numbers only";
              }
            }}
          />
          <SubmitControl label="Create service" errorAlert={status} disabled={isSubmitting} />
        </Form>
      )}
    </Formik>
  );
};

export default CreateService;
