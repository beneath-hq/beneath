import clsx from "clsx";
import React, { FC } from "react";

import { makeStyles, Theme } from "@material-ui/core";
import Fade from "@material-ui/core/Fade";
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
  highlightedCell: {
    backgroundColor: "rgba(255, 255, 255, 0.08)",
  },
}));

export interface RecordsTableProps {
  schema: Schema;
  records: Records_records_data[] | null;
  highlightTopN?: number;
}

const RecordsTable: FC<RecordsTableProps> = ({ schema, records, highlightTopN }) => {
  const classes = useStyles();
  return (
    <div className={classes.paper}>
      <Table className={classes.table} size="small">
        <TableHead>
          <TableRow>{schema.columns.map((column) => column.makeTableHeaderCell(classes.cell))}</TableRow>
        </TableHead>
        <TableBody>
          {records &&
            records.map((record, idx) => (
              <TableRow
                key={record.recordID}
                className={clsx(classes.row, idx < (highlightTopN || 0) && classes.highlightedCell)}
                hover={true}
              >
                {schema.columns.map((column) => column.makeTableCell(record.data, classes.cell))}
              </TableRow>
            ))}
        </TableBody>
      </Table>
    </div>
  );
};

export default RecordsTable;
