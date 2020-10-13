import { useMutation } from "@apollo/client";
import { Field, Formik } from "formik";
import React, { FC } from "react";

import { QUERY_STREAM_INSTANCES, STAGE_STREAM_INSTANCE } from "../../apollo/queries/stream";
import { StageStreamInstance, StageStreamInstanceVariables } from "../../apollo/types/StageStreamInstance";
import { Form, handleSubmitMutation, SelectField as FormikSelectField, TextField as FormikTextField } from "../formik";
import SubmitControl from "../forms/SubmitControl";
import FormikRadioGroup from "components/formik/RadioGroup";
import { StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName } from "apollo/types/StreamByOrganizationProjectAndName";
import { Instance } from "pages/stream";

export interface CreateInstanceProps {
  stream: StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName;
  instances: Instance[];
  setInstance: (instance: Instance | null) => void;
  setOpenDialogID: (dialogID: "create" | "promote" | "delete" | null) => void;
}

const CreateInstance: FC<CreateInstanceProps> = ({ stream, instances, setInstance, setOpenDialogID }) => {
  const [stageStreamInstance] = useMutation<StageStreamInstance, StageStreamInstanceVariables>(STAGE_STREAM_INSTANCE, {
    onCompleted: (data) => {
      if (data?.stageStreamInstance) {
        setInstance(data.stageStreamInstance);
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
      // initialStatus={error?.message}
      initialValues={initialValues}
      onSubmit={async (values, actions) =>
        handleSubmitMutation(
          values,
          actions,
          stageStreamInstance({
            variables: {
              streamID: stream.streamID,
              version: values.version,
              makePrimary: values.makePrimary==="true",
            },
            update: (cache, { data }) => {
              if (data && data.stageStreamInstance) {
                const queryData = cache.readQuery({
                  query: QUERY_STREAM_INSTANCES,
                  variables: {organizationName: stream.project.organization.name, projectName: stream.project.name, streamName: stream.name },
                }) as any;

                let newInstanceList = [data.stageStreamInstance].concat(queryData.streamInstancesByOrganizationProjectAndStreamName);

                // if the instance was made primary, remove old instances
                if (data.stageStreamInstance.madePrimaryOn) {
                  newInstanceList = newInstanceList.filter(
                    (instnc: Instance) => instnc.version >= data.stageStreamInstance.version
                  );
                }

                cache.writeQuery({
                  query: QUERY_STREAM_INSTANCES,
                  variables: {organizationName: stream.project.organization.name, projectName: stream.project.name, streamName: stream.name },
                  data: {streamInstancesByOrganizationProjectAndStreamName: newInstanceList },
                });
              }
            },
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
