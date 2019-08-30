import clsx from "clsx";
import React, { FC } from "react";

import {
  Icon,
  makeStyles,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  Theme,
  Tooltip,
} from "@material-ui/core";
import InfoIcon from "@material-ui/icons/InfoSharp";

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
  headerCell: {
    padding: "6px 16px 6px 16px",
    whiteSpace: "nowrap",
  },
  headerCellContent: {
    display: "inline-flex",
  },
  headerCellText: {
    flex: "1",
  },
  headerCellInfo: {
    fontSize: theme.typography.body1.fontSize,
    marginLeft: "5px",
    marginTop: "1px",
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
          <TableRow>
            {schema.columns.map((column) => (
              <TableCell key={column.name} className={clsx(classes.cell, classes.headerCell)}>
                <div className={classes.headerCellContent}>
                  <span className={classes.headerCellText}>{column.displayName}</span>
                  {column.doc && (
                    <Tooltip title={column.doc} interactive>
                      <Icon className={classes.headerCellInfo}>
                        <InfoIcon fontSize="inherit" />
                      </Icon>
                    </Tooltip>
                  )}
                </div>
              </TableCell>
            ))}
          </TableRow>
        </TableHead>
        <TableBody>
          {records &&
            records.map((record, idx) => (
              <TableRow
                key={record.recordID}
                className={clsx(classes.row, idx < (highlightTopN || 0) && classes.highlightedCell)}
                hover={true}
              >
                {schema.columns.map((column) => (
                  <TableCell key={column.name} className={classes.cell} align={column.isNumeric() ? "right" : "left"}>
                    {column.formatRecord(record.data)}
                  </TableCell>
                ))}
              </TableRow>
            ))}
        </TableBody>
      </Table>
    </div>
  );
};

export default RecordsTable;
