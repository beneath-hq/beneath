import React, { FC } from "react";

import { makeStyles, Theme } from "@material-ui/core";
import Table from "@material-ui/core/Table";
import TableBody from "@material-ui/core/TableBody";
import TableHead from "@material-ui/core/TableHead";
import TableRow from "@material-ui/core/TableRow";

import { Records_records_data } from "../apollo/types/Records";
import { Schema } from "./stream/schema";

const useStyles = makeStyles((theme: Theme) => ({
  paper: {
    width: "100%",
    overflowX: "auto",
  },
  table: {},
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

export interface RecordsTableProps {
  schema: Schema;
  records: Records_records_data[] | null;
}

const RecordsTable: FC<RecordsTableProps> = ({ schema, records }) => {
  const classes = useStyles();
  return (
    <div className={classes.paper}>
      <Table className={classes.table} size="small">
        <TableHead>
          <TableRow>{schema.columns.map((column) => column.makeTableHeaderCell(classes.cell))}</TableRow>
        </TableHead>
        <TableBody>
          {records &&
            records.map((record) => (
              <TableRow key={record.recordID} className={classes.row} hover={true}>
                {schema.columns.map((column) => column.makeTableCell(record.data, classes.cell))}
              </TableRow>
            ))}
        </TableBody>
      </Table>
    </div>
  );
};

export default RecordsTable;
