import { useMutation, useQuery } from "@apollo/client";
import { Field, Formik } from "formik";
import { useRouter } from "next/router";
import React, { FC } from "react";

import { QUERY_PROJECTS_FOR_USER } from "../../apollo/queries/project";
import { STAGE_STREAM } from "../../apollo/queries/stream";
import { StreamSchemaKind } from "../../apollo/types/globalTypes";
import { ProjectsForUser, ProjectsForUserVariables } from "../../apollo/types/ProjectsForUser";
import { StageStream, StageStreamVariables } from "../../apollo/types/StageStream";
import useMe from "../../hooks/useMe";
import { toURLName, toBackendName } from "../../lib/names";
import { Form, handleSubmitMutation, SelectField as FormikSelectField, TextField as FormikTextField } from "../formik";
import SubmitControl from "../forms/SubmitControl";
import FormikCodeEditor from "components/formik/CodeEditor";
import { removeRedirectAfterAuth } from "lib/authRedirect";

interface Project {
  organization: { name: string };
  name: string;
  displayName?: string;
}

export interface CreateProjectProps {
  preselectedProject?: Project;
}

const INITIAL_SCHEMA = `# Streams in Beneath have a predefined schema, which includes one or more
# "key" fields that uniquely identify a record. The key is used for indexing
# and compaction.

# To create a stream from Python, see
# https://about.beneath.dev/docs/reading-writing-data/creating-streams/

# Example schema:

" Description of the stream goes here "
type Movie @stream @key(fields: ["title", "released_on"]) {
  title: String!
  released_on: Timestamp!
  director: String!
  platform: Platform!
  description: String # optional field (no '!' after the type)
}

enum Platform {
  Cinema
  Apple
  Amazon
  Disney
  Netflix
}
`;

const CreateStream: FC<CreateProjectProps> = ({ preselectedProject }) => {
  const me = useMe();
  const router = useRouter();
  const [stageStream] = useMutation<StageStream, StageStreamVariables>(STAGE_STREAM, {
    onCompleted: (data) => {
      if (data?.stageStream) {
        const orgName = toURLName(data.stageStream.project.organization.name);
        const projName = toURLName(data.stageStream.project.name);
        const streamName = toURLName(data.stageStream.name);
        const href = `/stream?organization_name=${orgName}&project_name=${projName}&stream_name=${streamName}`;
        const as = `/${orgName}/${projName}/${streamName}`;
        router.replace(href, as, { shallow: true });
      }
    },
  });

  const { data, loading, error } = useQuery<ProjectsForUser, ProjectsForUserVariables>(QUERY_PROJECTS_FOR_USER, {
    variables: { userID: me?.personalUserID || "" },
    skip: !me,
  });

  removeRedirectAfterAuth();

  const initialValues = {
    project: data?.projectsForUser?.length ? (preselectedProject ? preselectedProject : data.projectsForUser[0]) : null,
    name: "",
    schemaKind: StreamSchemaKind.GraphQL,
    schema: INITIAL_SCHEMA,
  };

  return (
    <Formik
      initialStatus={error?.message}
      initialValues={initialValues}
      onSubmit={async (values, actions) =>
        handleSubmitMutation(
          values,
          actions,
          stageStream({
            variables: {
              organizationName: toBackendName(values.project?.organization.name || ""),
              projectName: toBackendName(values.project?.name || ""),
              streamName: toBackendName(values.name),
              schemaKind: values.schemaKind,
              schema: values.schema,
              allowManualWrites: true,
            },
          })
        )
      }
    >
      {({ isSubmitting, status }) => (
        <Form title="Create stream">
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
                return "Stream names should be between 3 and 40 characters long";
              }
              if (!val.match(/^[_\-a-z][_\-a-z0-9]+$/)) {
                return "Stream names should consist of lowercase letters, numbers, underscores and dashes (cannot start with a number)";
              }
            }}
            component={FormikTextField}
            label="Name"
            required
          />
          <Field
            name="schemaKind"
            validate={(kind?: string) => {
              if (!kind) {
                return "Select a schema kind for the stream";
              }
            }}
            component={FormikSelectField}
            label="Schema language"
            required
            disableClearable
            options={[StreamSchemaKind.GraphQL]}
          />
          <Field
            name="schema"
            validate={(schema?: string) => {
              if (!schema) {
                return "You must provide a valid schema";
              }
            }}
            component={FormikCodeEditor}
            label="Schema"
            required
            multiline
            rows={20}
            language={"graphql"}
          />
          <SubmitControl label="Create stream" errorAlert={status} disabled={isSubmitting} />
        </Form>
      )}
    </Formik>
  );
};

export default CreateStream;
