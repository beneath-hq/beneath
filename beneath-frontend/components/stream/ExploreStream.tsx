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

import connection from "../../lib/connection";
import { QueryStream, QueryStream_stream } from "../../apollo/types/QueryStream";
import VSpace from "../VSpace";
import { Schema } from "./schema";

const useStyles = makeStyles((theme: Theme) => ({
  table: {
    width: "100%",
  },
}));

const ExploreStream: FC<QueryStream> = ({ stream }) => {
  const schema = new Schema(stream);

  const [data, setData] = useState({ records: [] });

  useEffect(() => {
    const fetchRecords = async () => {
      const url = `${connection.GATEWAY_URL}/projects/${stream.project.name}/streams/${stream.name}`;
      const res = await fetch(url);
      const json = await res.json();
      setData({
        records: data.records.concat(json)
      });
    };

    fetchRecords();
  }, []);

  const classes = useStyles();
  return (
    <Table className={classes.table} size="small">
      <TableHead>
        <TableRow>{schema.columns.map((column) => column.makeTableHeaderCell())}</TableRow>
      </TableHead>
      <TableBody>
        {data.records.map((record) => (
          <TableRow key={schema.makeUniqueIdentifier(record)}>
            {schema.columns.map((column) => column.makeTableCell(record))}
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
};

export default ExploreStream;
