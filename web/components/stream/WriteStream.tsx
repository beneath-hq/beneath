import { useMutation } from "@apollo/client";
import _ from "lodash";
import React, { FC } from "react";

import { Typography } from "@material-ui/core";

import { CREATE_RECORDS } from "../../apollo/queries/local/records";
import { CreateRecords, CreateRecordsVariables } from "../../apollo/types/CreateRecords";
import { StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName } from "../../apollo/types/StreamByOrganizationProjectAndName";
import { Schema } from "./schema";
import { Formik, Form, Field } from "formik";
import { handleSubmitMutation } from "components/formik";
import FormikTextField from "components/formik/TextField";
import SubmitControl from "components/forms/SubmitControl";
import { validateValue, getPlaceholder } from "./FilterField";

interface WriteStreamProps {
  stream: StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName;
}

const WriteStream: FC<WriteStreamProps> = ({ stream }) => {
  const schema = new Schema(stream);

  const [createRecords] = useMutation<CreateRecords, CreateRecordsVariables>(CREATE_RECORDS, {
    onCompleted: ({ createRecords }) => {
      if (createRecords.error) {
        console.log(createRecords.error);
      } else {
        console.log("Successfully wrote record, but it might take a while before it shows up.");
      }
    },
  });

  const initialValues: any = {
    data: {},
    instanceID: stream.primaryStreamInstanceID,
  };

  schema.columns.map((col) => {
    initialValues.data[col.name]= "";
  });

  return (
    <Formik
      initialValues={initialValues}
      onSubmit={async (values, actions) =>
        handleSubmitMutation(
          values,
          actions,
          createRecords({
            variables: {
              json: values.data,
              instanceID: values.instanceID,
            },
          })
        )
      }
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
                placeholder="PLACEHOLDER" // Placeholders aren't working. Need to do some Formik edits I think.
                // placeholder={getPlaceholder(col.fieldType)} // TODO: need to get a FieldType from the schema.Column class. Benjamin handling.
                validate={(val: string) => {
                  // return validateValue(col.fieldType, val) // TODO: need to get a FieldType from the schema.Column class. Benjamin handling.
                }}
                component={FormikTextField}
                label={col.displayName}
                // required={col.nullable} // TODO: waiting on Benjamin to add this to the Column class
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
