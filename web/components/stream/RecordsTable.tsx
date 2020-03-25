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
import SearchIcon from "@material-ui/icons/SearchSharp";

import { Record } from "./beneath";
import { Schema } from "./schema";

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
  records: Record[];
  showTimestamps?: boolean;
}

const RecordsTable: FC<RecordsTableProps> = ({ schema, records, showTimestamps }) => {
  const classes = useStyles();
  const columns = schema.getColumns(showTimestamps);
  return (
    <div className={classes.paper}>
      <Table className={classes.table} size="small">
        <TableHead>
          <TableRow>
            {columns.map((column) => (
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
                  {column.key && (
                    <Tooltip title={"Column is part of an index"} interactive>
                      <Icon className={classes.headerCellInfo}>
                        <SearchIcon fontSize="inherit" />
                      </Icon>
                    </Tooltip>
                  )}
                </div>
              </TableCell>
            ))}
          </TableRow>
        </TableHead>
        <TableBody>
          {records.map((record, idx) => (
            <TableRow
              key={`${record["@meta"].key};${record["@meta"].timestamp}`}
              className={clsx(classes.row, record["@meta"].flash && classes.highlightedCell)}
              hover={true}
            >
              {columns.map((column) => (
                <TableCell key={column.name} className={classes.cell} align={column.isNumeric() ? "right" : "left"}>
                  {column.formatRecord(record)}
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
