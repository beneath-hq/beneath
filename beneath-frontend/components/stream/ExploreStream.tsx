import React, { FC, useEffect, useState } from "react";
import Moment from "react-moment";

import { makeStyles, Theme } from "@material-ui/core";
import Button from "@material-ui/core/Button";
import FormGroup from "@material-ui/core/FormGroup";
import Table from "@material-ui/core/Table";
import TableBody from "@material-ui/core/TableBody";
import TableCell, { TableCellProps } from "@material-ui/core/TableCell";
import TableHead from "@material-ui/core/TableHead";
import TableRow from "@material-ui/core/TableRow";
import TextField from "@material-ui/core/TextField";
import Typography from "@material-ui/core/Typography";

import { Query } from "react-apollo";
import { QUERY_RECORDS } from "../../apollo/queries/local/records";
import connection from "../../lib/connection";
import { QueryStream, QueryStream_stream } from "../../apollo/types/QueryStream";
import { Records, RecordsVariables } from "../../apollo/types/Records";
import VSpace from "../VSpace";
import { Schema } from "./schema";
import Loading from "../Loading";

const useStyles = makeStyles((theme: Theme) => ({
  table: {
    width: "100%",
  },
}));

const ExploreStream: FC<QueryStream> = ({ stream }) => {
  const classes = useStyles();
  const schema = new Schema(stream);

  const vars: RecordsVariables = {
    projectName: stream.project.name,
    streamName: stream.name,
    keyFields: schema.keyFields,
    limit: 100,
    where: null,
  };

  return (
    <Query<Records, RecordsVariables> query={QUERY_RECORDS} variables={vars}>
      {({ loading, error, data }) => {
        if (loading) {
          return <Loading justify="center" />;
        }

        if (error || data === undefined) {
          return <p>Error: {JSON.stringify(error)}</p>;
        }

        return (
          <Table className={classes.table} size="small">
            <TableHead>
              <TableRow>{schema.columns.map((column) => column.makeTableHeaderCell())}</TableRow>
            </TableHead>
            <TableBody>
              {data.records.map((record) => (
              <TableRow key={record.recordID}>
                {schema.columns.map((column) => column.makeTableCell(record.data))}
              </TableRow>
            ))}
            </TableBody>
          </Table>
        );
      }}
    </Query>
  );
};

export default ExploreStream;


