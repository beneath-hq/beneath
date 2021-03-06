import _ from "lodash";
import React, { FC } from "react";
import { Button, Dialog, DialogContent, Typography } from "@material-ui/core";
import AddBoxIcon from "@material-ui/icons/AddBox";
import { Client } from "beneath";

import { TableByOrganizationProjectAndName_tableByOrganizationProjectAndName } from "../../apollo/types/TableByOrganizationProjectAndName";
import { Schema, validateValue, serializeValue } from "./schema";
import { Formik, Form, Field } from "formik";
import FormikTextField from "components/formik/TextField";
import SubmitControl from "components/forms/SubmitControl";
import { useToken } from "hooks/useToken";

interface WriteTableProps {
  table: TableByOrganizationProjectAndName_tableByOrganizationProjectAndName;
  instanceID: string;
  buttonClassName: string;
}

const WriteTable: FC<WriteTableProps> = ({ table: tableMetadata, instanceID, buttonClassName }) => {
  const schema = new Schema(tableMetadata.avroSchema, tableMetadata.tableIndexes);
  const token = useToken();
  const [writeDialog, setWriteDialog] = React.useState(false);

  const client = new Client({ secret: token || undefined });
  const table = client.findTable({ instanceID });

  const initialValues: any = {
    data: {},
    instanceID,
  };

  schema.columns.map((col) => {
    initialValues.data[col.name] = "";
  });

  return (
    <>
      <Button
        variant="outlined"
        onClick={() => setWriteDialog(true)}
        endIcon={<AddBoxIcon />}
        size="small"
        classes={{ root: buttonClassName }}
      >
        Write record
      </Button>
      <Dialog open={writeDialog} fullWidth={true} maxWidth={"xs"} onBackdropClick={() => setWriteDialog(false)}>
        <DialogContent>
          <Formik
            initialValues={initialValues}
            onSubmit={async (values, actions) => {
              // clear any previous error
              actions.setStatus(null);

              // serialize the values
              const data: { [key: string]: any } = {};
              schema.columns.map((col) => {
                data[col.name] = serializeValue(col.inputType, values.data[col.name], col.isNullable);
              });

              const { writeID, error } = await table.write([data]);
              if (error) {
                actions.setStatus(error.message);
              }
              if (writeID) {
                // successfully wrote record
                setWriteDialog(false);
              }
            }}
          >
            {({ isSubmitting, status }) => (
              <Form>
                <Typography component="h2" variant="h1" gutterBottom>
                  Write a record
                </Typography>
                {schema.columns.map((col, idx) => (
                  <Field
                    key={idx}
                    name={"data." + col.name}
                    // placeholder={getPlaceholder(col.inputType)} // placeholders aren't this easy
                    validate={(val: string) => {
                      return validateValue(col.inputType, val);
                    }}
                    component={FormikTextField}
                    label={`${col.displayName} (${col.typeName})`}
                    required={!col.isNullable}
                  />
                ))}
                <SubmitControl label="Write record" errorAlert={status} disabled={isSubmitting} />
              </Form>
            )}
          </Formik>
        </DialogContent>
      </Dialog>
    </>
  );
};

export default WriteTable;
