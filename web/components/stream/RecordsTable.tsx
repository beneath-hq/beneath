import { Record } from "beneath-react";
import clsx from "clsx";
import dynamic from "next/dynamic";
import React, { FC } from "react";
import times from 'lodash/times';

const Moment = dynamic(import("react-moment"), { ssr: false });

import {
  Button,
  makeStyles,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  Theme,
  Tooltip,
  Box, Typography, Paper,
  fade,
} from "@material-ui/core";
import InfoIcon from "@material-ui/icons/InfoSharp";

import { Schema } from "./schema";
import ContentContainer, { ContentContainerProps } from "../ContentContainer";

const useStyles = makeStyles((theme: Theme) => ({
  cell: {
    borderLeft: `1px solid ${theme.palette.border.paper}`,
    "&:first-child": {
      borderLeft: "none",
    },
  },
  keyCell: {
    backgroundColor: fade(theme.palette.background.default, 0.25),
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
    alignItems: "center"
  },
  headerCellText: {
    flex: "1",
    marginRight: theme.spacing(1),
    fontSize: theme.typography.pxToRem(16),
    fontWeight: "bold"
  },
  headerCellPaper: {
    marginLeft: theme.spacing(1),
    paddingLeft: theme.spacing(.75),
    paddingRight: theme.spacing(.75),
    cursor: "default",
    height: "18px",
  },
  headerCellPaperText: {
    display: "table-cell",
    textAlign: "center",
    verticalAlign: "middle",
    lineHeight: "18px",
  },
  headerCellPaperKey: {
    backgroundColor: theme.palette.primary.dark,
  },
  headerCellPaperType: {
    backgroundColor: theme.palette.border.background
  },
  headerCellPaperInfo: {
    backgroundColor: theme.palette.border.background,
    display: "flex",
    flexDirection: "column",
    justifyContent: "center",
  },
  headerCellInfo: {
    fontSize: theme.typography.h3.fontSize,
  },
  emptyCell: {
    height: theme.spacing(4)
  }
}));

export interface RecordsTableProps extends ContentContainerProps {
  schema?: Schema;
  records?: Record[];
  showTimestamps?: boolean;
  fetchMore?: () => void;
}

const RecordsTable: FC<RecordsTableProps> = ({
  schema,
  records,
  showTimestamps,
  fetchMore,
  loading,
  ...containerProps
}) => {
  const classes = useStyles();
  const columns = schema?.getColumns(showTimestamps);
  return (
    <>
      <ContentContainer {...containerProps}>
        {columns && (
          <Table size="small">
            <TableHead>
              <TableRow>
                {columns.map((column) => (
                  <TableCell key={column.name} className={clsx(classes.cell, classes.headerCell)}>
                    <div className={classes.headerCellContent}>
                      <span className={classes.headerCellText}>{column.displayName}</span>
                      {column.isKey && (
                        <Tooltip title={"Column is part of an index"} interactive>
                          <Paper className={clsx(classes.headerCellPaper, classes.headerCellPaperKey)}>
                            <Typography variant="caption" className={classes.headerCellPaperText}>Key</Typography>
                          </Paper>
                        </Tooltip>
                      )}
                      {column.displayName !== "Time ago" && (
                        <Tooltip title={column.typeDescription} interactive>
                          <Paper className={clsx(classes.headerCellPaper, classes.headerCellPaperType)}>
                            <Typography variant="caption" className={classes.headerCellPaperText}>{column.typeName}</Typography>
                          </Paper>
                        </Tooltip>
                      )}
                      {column.doc && (
                        <Tooltip title={column.doc} interactive>
                          <Paper className={clsx(classes.headerCellPaper, classes.headerCellPaperInfo)}>
                            <InfoIcon className={classes.headerCellInfo}/>
                          </Paper>
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
                    className={clsx(record["@meta"].flash && classes.highlightedCell)}
                    hover={true}
                  >
                    {columns.map((column) => (
                          <TableCell
                            key={column.name}
                            className={clsx(classes.cell, column.isKey && classes.keyCell)}
                            align={column.isNumeric ? "right" : "left"}
                          >
                            {column.type === "timeago" ? (
                              <Moment fromNow ago date={column.formatRecord(record)} />
                            ) : (
                                column.formatRecord(record)
                              )}
                          </TableCell>
                        )
                    )}
                  </TableRow>
                ))}
                {records && records.length === 0 && times(3, (num) => (
                  <TableRow key={num} hover={true}>
                    {columns.map((column, idx) => (
                      <TableCell
                        key={idx}
                        className={clsx(classes.cell, column.isKey && classes.keyCell, classes.emptyCell)}
                        align={column.isNumeric ? "right" : "left"}
                      >
                      </TableCell>
                    ))}
                  </TableRow>
                ))}
            </TableBody>
          </Table>
        )}
      </ContentContainer>
      {fetchMore && (
        <Box mt={2} display="flex" justifyContent="center">
          <Button variant="contained" color="primary" disabled={loading} onClick={fetchMore}>
            Fetch more
          </Button>
        </Box>
      )}
    </>
  );
};

export default RecordsTable;
