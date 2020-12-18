import { useMutation, useQuery } from "@apollo/client";
import { Field, Formik } from "formik";
import React, { FC } from "react";

import { Form, handleSubmitMutation, TextField as FormikTextField } from "components/formik";
import SubmitControl from "components/forms/SubmitControl";
import FormikSelectField from "components/formik/SelectField";
import useMe from "hooks/useMe";
import { StreamsForUser, StreamsForUserVariables } from "apollo/types/StreamsForUser";
import { QUERY_STREAMS_FOR_USER } from "apollo/queries/stream";
import { UpdateServiceStreamPermissions, UpdateServiceStreamPermissionsVariables } from "apollo/types/UpdateServiceStreamPermissions";
import { UPDATE_SERVICE_STREAM_PERMISSIONS } from "apollo/queries/service";
import { toURLName } from "lib/names";
import FormikRadioGroup from "components/formik/RadioGroup";

interface Stream {
  streamID: string;
  name: string;
  project: { name: string, organization: { name: string } };
}

export interface Props {
  serviceID: string;
  onCompleted: () => void;
}

const AddPermission: FC<Props> = ({ serviceID, onCompleted }) => {
  const me = useMe();
  const { data, loading, error } = useQuery<StreamsForUser, StreamsForUserVariables>(QUERY_STREAMS_FOR_USER, {
    variables: { userID: me?.personalUserID || "" },
    skip: !me,
  });

  const [updateServiceStreamPermissions] = useMutation<UpdateServiceStreamPermissions, UpdateServiceStreamPermissionsVariables>(
    UPDATE_SERVICE_STREAM_PERMISSIONS, {
      onCompleted: (data) => {
        if (data.updateServiceStreamPermissions) {
          onCompleted()
        }
      }
    }
  );

  const initialValues = {
    stream: null as (Stream | null),
    read: "",
    write: "",
  };

  return (
    <>
      <Formik
        initialValues={initialValues}
        onSubmit={(values, actions) =>
          handleSubmitMutation(
            values,
            actions,
            updateServiceStreamPermissions({
              variables: {
                serviceID: serviceID,
                streamID: values.stream?.streamID as string,
                read: values.read === "true" ? true : false,
                write: values.write === "true" ? true : false,
              },
            })
          )
        }
      >
        {({ isSubmitting, status }) => (
          <Form
            title="Add permission"
            variant="embedded"
          >
            <Field
              name="stream"
              validate={(stream?: Stream) => {
                if (!stream) {
                  return "Select a stream";
                }
              }}
              component={FormikSelectField}
              label="Stream"
              required
              loading={loading}
              options={data?.streamsForUser || []}
              getOptionLabel={(option: Stream) => `${toURLName(option.project.organization.name)}/${toURLName(option.project.name)}/${toURLName(option.name)}`}
              getOptionSelected={(option: Stream, value: Stream) => {
                return option.name === value.name;
              }}
            />
            <Field
              name="read"
              component={FormikRadioGroup}
              label="Read access"
              required
              validate={(value: string) => {
                if (value === "") {
                  return "Select an option"
                }
              }}
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
              validate={(value: string) => {
                if (value === "") {
                  return "Select an option"
                }
              }}
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
