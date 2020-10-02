import { Record } from "beneath-react";
import clsx from "clsx";
import dynamic from "next/dynamic";
import React, { FC } from "react";
const Moment = dynamic(import("react-moment"), { ssr: false });

import {
  Button,
  Grid,
  Icon,
  makeStyles,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  Theme,
  Tooltip,
  Typography,
  Paper,
} from "@material-ui/core";
import InfoIcon from "@material-ui/icons/InfoSharp";
import SearchIcon from "@material-ui/icons/SearchSharp";

import { Schema } from "./schema";

const useStyles = makeStyles((theme: Theme) => ({
  paper: {
    width: "100%",
    overflowX: "auto",
  },
  table: {},
  tableHead: {
    backgroundColor: theme.palette.background.medium,
  },
  row: {
    "&:last-child": {
      "& td": {
        borderBottom: "none",
      },
    },
  },
  cell: {
    borderBottom: `1px solid ${theme.palette.border.paper}`,
    borderLeft: `1px solid ${theme.palette.border.paper}`,
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
  message: {
    margin: "5rem 0",
  },
}));

export interface RecordsTableProps {
  schema?: Schema;
  records?: Record[];
  showTimestamps?: boolean;
  loading?: boolean;
  error?: boolean;
  message?: string;
  fetchMore?: () => void;
}

const RecordsTable: FC<RecordsTableProps> = ({ schema, records, showTimestamps, loading, error, message, fetchMore }) => {
  const classes = useStyles();
  const columns = schema?.getColumns(showTimestamps);
  return (
    <Grid container direction="column" spacing={2}>
      <Grid item xs={12}>
        <Paper variant="outlined" className={classes.paper}>
          {columns && (
            <Table className={classes.table} size="small">
              <TableHead className={classes.tableHead}>
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
                        {column.isKey && (
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
                {records &&
                  records.map((record, _) => (
                    <TableRow
                      key={`${record["@meta"].key};${record["@meta"].timestamp}`}
                      className={clsx(classes.row, record["@meta"].flash && classes.highlightedCell)}
                      hover={true}
                    >
                      {columns.map((column) => (
                        <TableCell
                          key={column.name}
                          className={classes.cell}
                          align={column.isNumeric ? "right" : "left"}
                        >
                          {column.type === "timeago" ? (
                            <Moment fromNow ago date={column.formatRecord(record)} />
                          ) : (
                            column.formatRecord(record)
                          )}
                        </TableCell>
                      ))}
                    </TableRow>
                  ))}
              </TableBody>
            </Table>
          )}
          {message && (
            <div className={classes.message}>
              <Typography color={error ? "error" : "textSecondary"} align="center">
                {message}
              </Typography>
            </div>
          )}
        </Paper>
      </Grid>
      <Grid item container justify="center">
        {fetchMore && (
          <Grid item>
            <Button variant="contained" color="primary" disabled={loading} onClick={fetchMore}>
              Fetch more
            </Button>
          </Grid>
        )}
      </Grid>
    </Grid>
  );
};

export default RecordsTable;
