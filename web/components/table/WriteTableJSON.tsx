import { useMutation } from "@apollo/client";
import _ from "lodash";
import React, { FC } from "react";

import { Button, makeStyles, Theme, Typography } from "@material-ui/core";

import { CREATE_RECORDS } from "../../apollo/queries/local/records";
import { CreateRecords, CreateRecordsVariables } from "../../apollo/types/CreateRecords";
import { TableByOrganizationProjectAndName_tableByOrganizationProjectAndName } from "../../apollo/types/TableByOrganizationProjectAndName";
import CodeTextField from "../CodeTextField";
import { Schema } from "./schema";

const useStyles = makeStyles((theme: Theme) => ({
  paper: {
    width: "100%",
    overflowX: "auto",
  },
  submitButton: {
    marginTop: theme.spacing(3),
  },
  row: {
    "&:last-child": {
      "& td": {
        borderBottom: "none",
      },
    },
  },
  cell: {
    borderBottom: `1px solid ${theme.palette.divider}`,
    borderLeft: `1px solid ${theme.palette.divider}`,
    "&:first-child": {
      borderLeft: "none",
    },
  },
}));

interface WriteTableProps {
  table: TableByOrganizationProjectAndName_tableByOrganizationProjectAndName;
}

const WriteTable: FC<WriteTableProps> = ({ table }) => {
  // const schema = new Schema(table.avroSchema, table.tableIndexes);

  const [values, setValues] = React.useState({
    json: "",
    error: "",
    flash: "",
  } as { json: string; error: string | null; flash: string | null });

  const [createRecords, { loading, error }] = useMutation<CreateRecords, CreateRecordsVariables>(CREATE_RECORDS, {
    onCompleted: ({ createRecords }) => {
      if (createRecords.error) {
        setValues({ ...values, error: createRecords.error });
      } else {
        setValues({
          json: "",
          error: "",
          flash: "Successfully wrote record, but it might take a while before it shows up.",
        });
      }
    },
  });

  const handleChange = (name: string) => (event: any) => {
    setValues({ ...values, [name]: event.target.value });
  };

  const errorMsg = error ? error.message : values.error;

  const onSubmit = (e: any) => {
    e.preventDefault();
    const json = JSON.parse(values.json);
    createRecords({
      variables: {
        json,
        instanceID: table.primaryTableInstanceID as string,
      },
    });
  };

  const classes = useStyles();
  return (
    <form onSubmit={onSubmit}>
      <CodeTextField
        id="json"
        label="JSON"
        value={values.json}
        margin="normal"
        multiline
        rows={15}
        rowsMax={1000}
        fullWidth
        required
        onChange={handleChange("json")}
        helperText={<Typography variant="caption">Enter the new record as a JSON object</Typography>}
        errorText={errorMsg ? errorMsg : undefined}
        successText={values.flash ? values.flash : undefined}
      />
      <Button
        type="submit"
        variant="outlined"
        color="primary"
        className={classes.submitButton}
        disabled={loading || !isJSON(values.json)}
      >
        Save record
      </Button>
    </form>
  );
};

export default WriteTable;

const isJSON = (val: string): boolean => {
  try {
    JSON.parse(val);
    return true;
  } catch {
    return false;
  }
};
