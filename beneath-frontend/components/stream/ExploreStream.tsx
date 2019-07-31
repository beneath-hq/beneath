import React, { FC, useEffect, useState } from "react";
import { Query } from "react-apollo";

import { makeStyles, Theme } from "@material-ui/core";
import Button from "@material-ui/core/Button";
import Grid from "@material-ui/core/Grid";
import Table from "@material-ui/core/Table";
import TableBody from "@material-ui/core/TableBody";
import TableHead from "@material-ui/core/TableHead";
import TableRow from "@material-ui/core/TableRow";
import TextField from "@material-ui/core/TextField";
import Typography from "@material-ui/core/Typography";

import { QUERY_RECORDS } from "../../apollo/queries/local/records";
import { QueryStream } from "../../apollo/types/QueryStream";
import { Records, RecordsVariables } from "../../apollo/types/Records";
import Loading from "../Loading";
import VSpace from "../VSpace";
import { Schema } from "./schema";

const useStyles = makeStyles((theme: Theme) => ({
  table: {
    width: "100%",
  },
  submitButton: {
    marginTop: theme.spacing(3),
  },
  cell: {
    "borderBottom": `1px solid ${theme.palette.divider}`,
    "borderLeft": `1px solid ${theme.palette.divider}`,
    "&:first-child": {
      borderLeft: "none",
    },
  },
}));

const ExploreStream: FC<QueryStream> = ({ stream }) => {
  const schema = new Schema(stream);

  const [values, setValues] = React.useState({
    where: "",
    vars: {
      projectName: stream.project.name,
      streamName: stream.name,
      keyFields: schema.keyFields,
      limit: 100,
      where: "",
    } as RecordsVariables,
  });

  const handleChange = (name: string) => (event: any) => {
    setValues({ ...values, [name]: event.target.value });
  };

  const classes = useStyles();
  return (
    <Query<Records, RecordsVariables> query={QUERY_RECORDS} variables={values.vars}>
      {({ loading, error, data }) => {
        const errorMsg = error ? error.message : data ? data.records.error : null;

        const formElem = (
          <form
            onSubmit={(e) => {
              e.preventDefault();
              const where = values.where ? JSON.parse(values.where) : "";
              setValues({ ...values, vars: { ...values.vars, where } });
            }}
          >
            <Grid container spacing={2}>
              <Grid item xs>
                <TextField
                  id="where"
                  label="Query"
                  value={values.where}
                  margin="none"
                  onChange={handleChange("where")}
                  helperText={
                    <>
                      Query the stream on indexed fields
                      {errorMsg && (
                        <>
                          <br /><br />
                          <Typography variant="caption" color="error">
                            Error: {errorMsg}
                          </Typography>
                        </>
                      )}
                    </>
                  }
                  fullWidth
                />
              </Grid>
              <Grid item>
                <Button
                  type="submit"
                  variant="outlined"
                  color="primary"
                  className={classes.submitButton}
                  disabled={
                    loading ||
                    !(isJSON(values.where) || values.where.length === 0) ||
                    !(values.where.length <= 1024)
                  }
                >
                  Load
                </Button>
              </Grid>
            </Grid>
          </form>
        );

        let tableElem = null;
        if (loading) {
          tableElem = <Loading justify="center" />;
        } else {
          tableElem = (
            <Table className={classes.table} size="small">
              <TableHead>
                <TableRow>{schema.columns.map((column) => column.makeTableHeaderCell(classes.cell))}</TableRow>
              </TableHead>
              <TableBody>
                {data &&
                  data.records.data &&
                  data.records.data.map((record) => (
                    <TableRow key={record.recordID} hover={true}>
                      {schema.columns.map((column) => column.makeTableCell(record.data, classes.cell))}
                    </TableRow>
                  ))}
              </TableBody>
            </Table>
          );
        }

        return (
          <>
            {formElem}
            <VSpace units={2} />
            {tableElem}
          </>
        );
      }}
    </Query>
  );
};

export default ExploreStream;

const isJSON = (val: string): boolean => {
  try {
    JSON.parse(val);
    return true;
  } catch {
    return false;
  }
}
