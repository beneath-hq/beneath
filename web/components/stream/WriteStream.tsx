import _ from "lodash";
import React, { FC } from "react";
import { Typography } from "@material-ui/core";
import { BrowserClient } from "beneath";

import { StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName } from "../../apollo/types/StreamByOrganizationProjectAndName";
import { Schema } from "./schema";
import { Formik, Form, Field } from "formik";
import FormikTextField from "components/formik/TextField";
import SubmitControl from "components/forms/SubmitControl";
import { validateValue } from "./FilterField";
import { useToken } from "hooks/useToken";

interface WriteStreamProps {
  stream: StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName;
  instanceID: string;
  setWriteDialog: (writeDialog: boolean) => void;
}

const WriteStream: FC<WriteStreamProps> = ({ stream: streamMetadata, instanceID, setWriteDialog }) => {
  const schema = new Schema(streamMetadata.avroSchema, streamMetadata.streamIndexes);
  const token = useToken();

  const client = new BrowserClient({ secret: token || undefined });
  const stream = client.findStream({instanceID});

  const initialValues: any = {
    data: {},
    instanceID,
  };

  schema.columns.map((col) => {
    initialValues.data[col.name]= "";
  });

  const sanitize = (schema: Schema, data: any) => {
    // when a field is optional or a key, and an empty string, remove the field from the record object
    schema.columns.map((col) => {
      if ((col.isNullable || col.isKey) && data[col.name] === "") {
        delete data[col.name];
      }
    });

    return;
  };

  return (
    <Formik
      initialValues={initialValues}
      onSubmit={async (values, actions) => {
        sanitize(schema, values.data);
        const { writeID, error } = await stream.write(values.data);
        if (error) {
          const json = JSON.parse(error.message);
          actions.setStatus(json.error);
        } else {
          // clear previous error
          actions.setStatus("");
        }
        if (writeID) {
          console.log("Successfully wrote record, but it might take a while before it shows up.");
          setWriteDialog(false);
        }
      }}
    >
      {({ isSubmitting, status }) => (
        <Form>
          <Typography component="h2" variant="h1" gutterBottom>
            Write a record
          </Typography>
          {schema.columns.map((col, idx) => {
            return (
              <Field
                key={idx}
                name={"data." + col.name}
                // placeholder={getPlaceholder(col.inputType)} // placeholders aren't this easy
                validate={(val: string) => {
                  return validateValue(col.inputType, val);
                }}
                component={FormikTextField}
                label={col.displayName}
                required={!col.isNullable}
              />
            );}
          )}
          <SubmitControl label="Write record" errorAlert={status} disabled={isSubmitting} />
        </Form>
      )}
    </Formik>
  );
};

export default WriteStream;
