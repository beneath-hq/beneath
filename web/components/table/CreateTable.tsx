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
import { CREATE_TABLE } from "../../apollo/queries/table";
import { TableSchemaKind } from "../../apollo/types/globalTypes";
import { ProjectsForUser, ProjectsForUserVariables } from "../../apollo/types/ProjectsForUser";
import { CreateTable, CreateTableVariables } from "../../apollo/types/CreateTable";
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
import { makeTableAs, makeTableHref } from "./urls";

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

const CreateTableView: FC<Props> = ({ preselectedProject }) => {
  const me = useMe();
  const router = useRouter();
  const classes = useStyles();
  const [examplesDialog, setExamplesDialog] = useState(false);
  const [createTable] = useMutation<CreateTable, CreateTableVariables>(CREATE_TABLE, {
    onCompleted: (data) => {
      if (data?.createTable) {
        router.replace(makeTableHref(data.createTable), makeTableAs(data.createTable), { shallow: true });
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
    schemaKind: TableSchemaKind.GraphQL,
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
          createTable({
            variables: {
              input: {
                organizationName: toBackendName(values.project?.organization.name || ""),
                projectName: toBackendName(values.project?.name || ""),
                tableName: toBackendName(values.name),
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
        <Form title="Create table">
          <Typography gutterBottom>
            Tables are how you store data in Beneath. You can replay and subscribe to tables, run OLAP queries with SQL,
            and look up specific records by key. Learn more about tables in the{" "}
            <Link href="https://about.beneath.dev/docs/concepts/">Concepts</Link> docs.
          </Typography>
          <Typography>
            You may also be interested in:
            <ul>
              <li>
                <Link href="https://about.beneath.dev/docs/quick-starts/create-table/">
                  How to create a real-time table from Python
                </Link>
              </li>
              <li>
                <Link href="https://about.beneath.dev/docs/quick-starts/publish-pandas-dataframe/">
                  How to publish a Pandas DataFrame
                </Link>
              </li>
            </ul>
          </Typography>
          {/* TEMPORARY: Hide this dropdown while there's only one option */}
          {/* <Grid container>
            <Grid item xs={6} sm={4} md={3}>
              <Field
                name="schemaKind"
                validate={(kind?: string) => {
                  if (!kind) {
                    return "Select a schema kind for the table";
                  }
                }}
                component={FormikSelectField}
                label="Schema language"
                required
                disableClearable
                options={[TableSchemaKind.GraphQL]}
              />
            </Grid>
          </Grid> */}
          <Field
            name="schema"
            component={FormikCodeEditor}
            label="GraphQL schema"
            required
            multiline
            rows={20}
            language={"graphql"}
            helperText={
              <>
                Tables need a predefined schema. Check out the{" "}
                <Link href="https://about.beneath.dev/docs/reading-writing-data/schema-definition/" target="_blank">
                  schema docs
                </Link>{" "}
                or start with an{" "}
                <Link onClick={() => setExamplesDialog(true)} className={classes.link}>
                  example schema
                </Link>
                .
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
                    return "Select a project for the table";
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
                    return "Table names should be between 3 and 40 characters long";
                  }
                  if (!val.match(/^[_\-a-z][_\-a-z0-9]+$/)) {
                    return "Table names should consist of lowercase letters, numbers, underscores and dashes (cannot start with a number)";
                  }
                }}
                component={FormikTextField}
                label="Table name"
                required
              />
            </Grid>
          </Grid>
          <Typography>
            URL preview: https://beneath.dev/
            {values.project?.organization.name ? toURLName(values.project?.organization.name) : "USERNAME"}/
            {values.project?.name ? toURLName(values.project?.name) : "PROJECT"}/table:
            {values.name ? toURLName(values.name) : ""}
          </Typography>
          <VSpace units={3} />
          <Collapse title="Advanced settings">
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
          <SubmitControl
            label="Create table"
            errorAlert={status}
            disabled={isSubmitting || !values.schema || !values.name}
          />
        </Form>
      )}
    </Formik>
  );
};

export default CreateTableView;
