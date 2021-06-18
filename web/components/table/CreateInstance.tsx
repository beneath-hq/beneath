import { useMutation } from "@apollo/client";
import { Field, Formik } from "formik";
import { useRouter } from "next/router";
import React, { FC } from "react";

import { QUERY_STREAM_INSTANCES, CREATE_STREAM_INSTANCE, QUERY_STREAM } from "../../apollo/queries/table";
import { CreateTableInstance, CreateTableInstanceVariables } from "../../apollo/types/CreateTableInstance";
import { Form, handleSubmitMutation, TextField as FormikTextField } from "../formik";
import SubmitControl from "../forms/SubmitControl";
import FormikRadioGroup from "components/formik/RadioGroup";
import { TableByOrganizationProjectAndName_tableByOrganizationProjectAndName } from "apollo/types/TableByOrganizationProjectAndName";
import { TableInstance } from "components/table/types";
import { makeTableAs, makeTableHref } from "./urls";

export interface CreateInstanceProps {
  table: TableByOrganizationProjectAndName_tableByOrganizationProjectAndName;
  instances: TableInstance[];
  setOpenDialogID: (dialogID: "create" | "promote" | "delete" | null) => void;
}

const CreateInstance: FC<CreateInstanceProps> = ({ table, instances, setOpenDialogID }) => {
  const router = useRouter();
  const [createTableInstance] = useMutation<CreateTableInstance, CreateTableInstanceVariables>(CREATE_STREAM_INSTANCE, {
    onCompleted: (data) => {
      if (data?.createTableInstance) {
        router.replace(makeTableHref(table, data.createTableInstance), makeTableAs(table, data.createTableInstance));
      }
      setOpenDialogID(null);
    },
  });

  const highestVersion = instances[0] ? instances[0].version : -1;
  const newVersionSuggestion = highestVersion + 1;

  const initialValues = {
    version: newVersionSuggestion,
    makePrimary: "false",
  };

  return (
    <Formik
      initialValues={initialValues}
      onSubmit={async (values, actions) =>
        handleSubmitMutation(
          values,
          actions,
          createTableInstance({
            variables: {
              input: {
                tableID: table.tableID,
                version: values.version,
                makePrimary: values.makePrimary === "true",
              },
            },
            refetchQueries: [
              {
                query: QUERY_STREAM,
                variables: {
                  organizationName: table.project.organization.name,
                  projectName: table.project.name,
                  tableName: table.name,
                },
              },
              // TODO: instead, we should use Apollo's "update()" to update the cache for the list of instances
              {
                query: QUERY_STREAM_INSTANCES,
                variables: {
                  organizationName: table.project.organization.name,
                  projectName: table.project.name,
                  tableName: table.name,
                },
              },
            ],
          })
        )
      }
    >
      {({ isSubmitting, status }) => (
        <Form title="Create instance">
          <Field
            name="version"
            validate={(version: number) => {
              if (version <= highestVersion) {
                return `A version ${highestVersion} exists. New versions must be greater than existing versions.`;
              }
            }}
            component={FormikTextField}
            label="Version"
            required
          />
          <Field
            name="makePrimary"
            label="Make Primary"
            helperText="Make this instance primary if you are positive that you will not need previous instances. All previous instances will be deleted and you will not be able to recover the data."
            component={FormikRadioGroup}
            required
            options={[
              { value: "true", label: "True" },
              { value: "false", label: "False" },
            ]}
            row
          />
          <SubmitControl label="Create instance" errorAlert={status} disabled={isSubmitting} />
        </Form>
      )}
    </Formik>
  );
};

export default CreateInstance;
