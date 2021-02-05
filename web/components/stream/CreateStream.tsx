import { useMutation, useQuery } from "@apollo/client";
import { Field, Formik } from "formik";
import { useRouter } from "next/router";
import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Grid,
  Link,
  List,
  ListItem,
  ListItemText,
  makeStyles,
  Typography,
} from "@material-ui/core";
import React, { FC, useState } from "react";

import { QUERY_PROJECTS_FOR_USER } from "../../apollo/queries/project";
import { CREATE_STREAM } from "../../apollo/queries/stream";
import { StreamSchemaKind } from "../../apollo/types/globalTypes";
import { ProjectsForUser, ProjectsForUserVariables } from "../../apollo/types/ProjectsForUser";
import { CreateStream, CreateStreamVariables } from "../../apollo/types/CreateStream";
import useMe from "../../hooks/useMe";
import { toURLName, toBackendName } from "../../lib/names";
import { Form, handleSubmitMutation, SelectField as FormikSelectField, TextField as FormikTextField } from "../formik";
import SubmitControl from "../forms/SubmitControl";
import FormikCodeEditor from "components/formik/CodeEditor";
import { removeRedirectAfterAuth } from "lib/authRedirect";
import FormikCheckbox from "components/formik/Checkbox";
import FormikRadioGroup from "components/formik/RadioGroup";
import VSpace from "components/VSpace";
import Collapse from "components/Collapse";
import EXAMPLE_SCHEMAS from "lib/exampleSchemas";
import SchemaEditorFooter from "./SchemaEditorFooter";
import { makeStreamAs, makeStreamHref } from "./urls";

const useStyles = makeStyles((theme) => ({
  list: {
    width: "100%",
    paddingTop: theme.spacing(0),
    paddingBottom: theme.spacing(0),
  },
  link: {
    cursor: "pointer",
  },
}));

interface Project {
  organization: { name: string };
  name: string;
  displayName?: string;
}

interface Props {
  preselectedProject?: Project;
}

const CreateStreamView: FC<Props> = ({ preselectedProject }) => {
  const me = useMe();
  const router = useRouter();
  const classes = useStyles();
  const [examplesDialog, setExamplesDialog] = useState(false);
  const [showAdvancedSettings, setShowAdvancedSettings] = useState(false);
  const [createStream] = useMutation<CreateStream, CreateStreamVariables>(CREATE_STREAM, {
    onCompleted: (data) => {
      if (data?.createStream) {
        router.replace(makeStreamHref(data.createStream), makeStreamAs(data.createStream), { shallow: true });
      }
    },
  });

  const { data, loading, error } = useQuery<ProjectsForUser, ProjectsForUserVariables>(QUERY_PROJECTS_FOR_USER, {
    variables: { userID: me?.personalUserID || "" },
    skip: !me,
  });

  if (typeof window !== "undefined") {
    removeRedirectAfterAuth();
  }

  const initialValues = {
    project: data?.projectsForUser?.length ? (preselectedProject ? preselectedProject : data.projectsForUser[0]) : null,
    name: "",
    schemaKind: StreamSchemaKind.GraphQL,
    schema: "",
    useLog: true,
    useIndex: true,
    useWarehouse: true,
    isLogRetentionFinite: "false",
    isIndexRetentionFinite: "false",
    isWarehouseRetentionFinite: "false",
    logRetentionHours: "",
    indexRetentionHours: "",
    warehouseRetentionHours: "",
  };

  return (
    <Formik
      initialStatus={error?.message}
      initialValues={initialValues}
      onSubmit={async (values, actions) =>
        handleSubmitMutation(
          values,
          actions,
          createStream({
            variables: {
              input: {
                organizationName: toBackendName(values.project?.organization.name || ""),
                projectName: toBackendName(values.project?.name || ""),
                streamName: toBackendName(values.name),
                schemaKind: values.schemaKind,
                schema: values.schema,
                allowManualWrites: true,
                useLog: values.useLog,
                useIndex: values.useIndex,
                useWarehouse: values.useWarehouse,
                logRetentionSeconds:
                  values.logRetentionHours !== "" ? Number(values.logRetentionHours) * 60 * 60 : null,
                indexRetentionSeconds:
                  values.indexRetentionHours !== "" ? Number(values.indexRetentionHours) * 60 * 60 : null,
                warehouseRetentionSeconds:
                  values.warehouseRetentionHours !== "" ? Number(values.warehouseRetentionHours) * 60 * 60 : null,
              },
            },
          })
        )
      }
    >
      {({ values, setFieldValue, isSubmitting, status }) => (
        <Form title="Create stream">
          <Typography>
            Streams in Beneath have a predefined schema. Submit this form and you'll be ready to write data to your
            stream.
          </Typography>
          {/* TEMPORARY: Hide this dropdown while there's only one option */}
          {/* <Grid container>
            <Grid item xs={6} sm={4} md={3}>
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
            </Grid>
          </Grid> */}
          <Field
            name="schema"
            validate={(schema?: string) => {
              if (!schema) {
                return "You must provide a valid schema";
              }
            }}
            component={FormikCodeEditor}
            label="GraphQL schema"
            required
            multiline
            rows={20}
            language={"graphql"}
            helperText={
              <>
                Check out the{" "}
                <Link href="https://about.beneath.dev/docs/reading-writing-data/schema-definition/">schema docs</Link>{" "}
                or start with an{" "}
                <Link onClick={() => setExamplesDialog(true)} className={classes.link}>
                  example schema
                </Link>
              </>
            }
          />
          <SchemaEditorFooter schemaKind={values.schemaKind} schema={values.schema} />
          <Dialog open={examplesDialog} onBackdropClick={() => setExamplesDialog(false)} maxWidth="xs" fullWidth>
            <DialogTitle>
              Load example schema
              <Typography variant="body2">Select a template to modify</Typography>
            </DialogTitle>
            <DialogContent>
              <List className={classes.list}>
                {EXAMPLE_SCHEMAS.map((exampleSchema) => (
                  <ListItem
                    key={exampleSchema.name}
                    button
                    onClick={() => {
                      setFieldValue("schema", exampleSchema.schema);
                      setExamplesDialog(false);
                    }}
                  >
                    <ListItemText primary={exampleSchema.name} secondary={exampleSchema.language} />
                  </ListItem>
                ))}
              </List>
            </DialogContent>
            <DialogActions>
              <Button onClick={() => setExamplesDialog(false)}>Close</Button>
            </DialogActions>
          </Dialog>
          <Grid container spacing={2}>
            <Grid item xs={12} sm={6}>
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
            </Grid>
            <Grid item xs={12} sm={6}>
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
                label="Stream name"
                required
              />
            </Grid>
          </Grid>
          <Typography>
            URL preview: https://beneath.dev/
            {values.project?.organization.name ? toURLName(values.project?.organization.name) : ""}/
            {values.project?.name ? toURLName(values.project?.name) : ""}/stream:
            {values.name ? toURLName(values.name) : ""}
          </Typography>
          <VSpace units={3} />
          <Collapse title="Advanced settings" in={showAdvancedSettings}>
            <Field name="useIndex" component={FormikCheckbox} type="checkbox" label="Enable indexing" />
            <Field name="useWarehouse" component={FormikCheckbox} type="checkbox" label="Enable SQL (warehouse)" />
            <Field
              name="isLogRetentionFinite"
              component={FormikRadioGroup}
              label="Log retention"
              required
              options={[
                { value: "false", label: "Infinite" },
                { value: "true", label: "Finite" },
              ]}
              row
            />
            {values.isLogRetentionFinite === "true" && (
              <Grid item xs={12} sm={6} md={3}>
                <Field
                  name="logRetentionHours"
                  component={FormikTextField}
                  label="Hours"
                  required
                  validate={(val: string) => {
                    if (!val.match(/^[0-9]*$/)) {
                      return "Numbers only";
                    }
                  }}
                />
              </Grid>
            )}
            {values.useIndex && (
              <>
                <Field
                  name="isIndexRetentionFinite"
                  component={FormikRadioGroup}
                  label="Index retention"
                  required
                  options={[
                    { value: "false", label: "Infinite" },
                    { value: "true", label: "Finite" },
                  ]}
                  row
                />
                {values.isIndexRetentionFinite === "true" && (
                  <Grid item xs={12} sm={6} md={3}>
                    <Field
                      name="indexRetentionHours"
                      component={FormikTextField}
                      label="Hours"
                      required
                      validate={(val: string) => {
                        if (!val.match(/^[0-9]*$/)) {
                          return "Numbers only";
                        }
                      }}
                    />
                  </Grid>
                )}
              </>
            )}
            {values.useWarehouse && (
              <>
                <Field
                  name="isWarehouseRetentionFinite"
                  component={FormikRadioGroup}
                  label="Warehouse retention"
                  required
                  options={[
                    { value: "false", label: "Infinite" },
                    { value: "true", label: "Finite" },
                  ]}
                  row
                />
                {values.isWarehouseRetentionFinite === "true" && (
                  <Grid item xs={12} sm={6} md={3}>
                    <Field
                      name="warehouseRetentionHours"
                      component={FormikTextField}
                      label="Hours"
                      required
                      validate={(val: string) => {
                        if (!val.match(/^[0-9]*$/)) {
                          return "Numbers only";
                        }
                      }}
                    />
                  </Grid>
                )}
              </>
            )}
          </Collapse>
          <SubmitControl label="Create stream" errorAlert={status} disabled={isSubmitting} />
        </Form>
      )}
    </Formik>
  );
};

export default CreateStreamView;
