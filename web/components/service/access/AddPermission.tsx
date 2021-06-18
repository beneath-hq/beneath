import { useMutation, useQuery } from "@apollo/client";
import { Field, Formik } from "formik";
import React, { FC } from "react";

import { Form, handleSubmitMutation, TextField as FormikTextField } from "components/formik";
import SubmitControl from "components/forms/SubmitControl";
import FormikSelectField from "components/formik/SelectField";
import useMe from "hooks/useMe";
import { TablesForUser, TablesForUserVariables } from "apollo/types/TablesForUser";
import { QUERY_STREAMS_FOR_USER } from "apollo/queries/table";
import {
  UpdateServiceTablePermissions,
  UpdateServiceTablePermissionsVariables,
} from "apollo/types/UpdateServiceTablePermissions";
import { QUERY_STREAM_PERMISSIONS_FOR_SERVICE, UPDATE_SERVICE_STREAM_PERMISSIONS } from "apollo/queries/service";
import { toURLName } from "lib/names";
import FormikRadioGroup from "components/formik/RadioGroup";

interface Table {
  tableID: string;
  name: string;
  project: { name: string; organization: { name: string } };
}

export interface Props {
  serviceID: string;
  onCompleted: () => void;
}

const AddPermission: FC<Props> = ({ serviceID, onCompleted }) => {
  const me = useMe();
  const { data, loading, error } = useQuery<TablesForUser, TablesForUserVariables>(QUERY_STREAMS_FOR_USER, {
    variables: { userID: me?.personalUserID || "" },
    skip: !me,
  });

  const [updateServiceTablePermissions] = useMutation<
    UpdateServiceTablePermissions,
    UpdateServiceTablePermissionsVariables
  >(UPDATE_SERVICE_STREAM_PERMISSIONS, {
    onCompleted: (data) => {
      if (data.updateServiceTablePermissions) {
        onCompleted();
      }
    },
  });

  const initialValues = {
    table: null as Table | null,
    read: "false",
    write: "false",
  };

  return (
    <>
      <Formik
        initialValues={initialValues}
        onSubmit={(values, actions) =>
          handleSubmitMutation(
            values,
            actions,
            updateServiceTablePermissions({
              variables: {
                serviceID: serviceID,
                tableID: values.table?.tableID as string,
                read: values.read === "true" ? true : false,
                write: values.write === "true" ? true : false,
              },
              refetchQueries: [{ query: QUERY_STREAM_PERMISSIONS_FOR_SERVICE, variables: { serviceID: serviceID } }],
            })
          )
        }
      >
        {({ isSubmitting, status }) => (
          <Form title="Add permission" variant="embedded">
            <Field
              name="table"
              validate={(table?: Table) => {
                if (!table) {
                  return "Select a table";
                }
              }}
              component={FormikSelectField}
              label="Table"
              required
              loading={loading}
              options={data?.tablesForUser || []}
              getOptionLabel={(option: Table) =>
                `${toURLName(option.project.organization.name)}/${toURLName(option.project.name)}/${toURLName(
                  option.name
                )}`
              }
              getOptionSelected={(option: Table, value: Table) => {
                return option.name === value.name;
              }}
            />
            <Field
              name="read"
              component={FormikRadioGroup}
              label="Read access"
              required
              options={[
                { value: "true", label: "True" },
                { value: "false", label: "False" },
              ]}
              row
            />
            <Field
              name="write"
              component={FormikRadioGroup}
              label="Write access"
              required
              options={[
                { value: "true", label: "True" },
                { value: "false", label: "False" },
              ]}
              row
            />
            <SubmitControl label="Add permission" errorAlert={status} disabled={isSubmitting} />
          </Form>
        )}
      </Formik>
    </>
  );
};

export default AddPermission;
