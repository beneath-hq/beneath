import { useMutation } from "@apollo/client";
import { Field, Formik } from "formik";
import { useRouter } from "next/router";
import React, { FC } from "react";

import { QUERY_STREAM_INSTANCES, CREATE_STREAM_INSTANCE, QUERY_STREAM } from "../../apollo/queries/stream";
import { CreateStreamInstance, CreateStreamInstanceVariables } from "../../apollo/types/CreateStreamInstance";
import { Form, handleSubmitMutation, TextField as FormikTextField } from "../formik";
import SubmitControl from "../forms/SubmitControl";
import FormikRadioGroup from "components/formik/RadioGroup";
import { StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName } from "apollo/types/StreamByOrganizationProjectAndName";
import { StreamInstance } from "components/stream/types";
import { makeStreamAs, makeStreamHref } from "./urls";

export interface CreateInstanceProps {
  stream: StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName;
  instances: StreamInstance[];
  setOpenDialogID: (dialogID: "create" | "promote" | "delete" | null) => void;
}

const CreateInstance: FC<CreateInstanceProps> = ({ stream, instances, setOpenDialogID }) => {
  const router = useRouter();
  const [createStreamInstance] = useMutation<CreateStreamInstance, CreateStreamInstanceVariables>(
    CREATE_STREAM_INSTANCE,
    {
      onCompleted: (data) => {
        if (data?.createStreamInstance) {
          router.replace(
            makeStreamHref(stream, data.createStreamInstance),
            makeStreamAs(stream, data.createStreamInstance)
          );
        }
        setOpenDialogID(null);
      },
    }
  );

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
          createStreamInstance({
            variables: {
              input: {
                streamID: stream.streamID,
                version: values.version,
                makePrimary: values.makePrimary === "true",
              },
            },
            refetchQueries: [
              {
                query: QUERY_STREAM,
                variables: {
                  organizationName: stream.project.organization.name,
                  projectName: stream.project.name,
                  streamName: stream.name,
                },
              },
              // TODO: instead, we should use Apollo's "update()" to update the cache for the list of instances
              {
                query: QUERY_STREAM_INSTANCES,
                variables: {
                  organizationName: stream.project.organization.name,
                  projectName: stream.project.name,
                  streamName: stream.name,
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
