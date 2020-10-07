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
  Box,
} from "@material-ui/core";
import InfoIcon from "@material-ui/icons/InfoSharp";
import SearchIcon from "@material-ui/icons/SearchSharp";

import { Schema } from "./schema";
import ContentContainer, { ContentContainerProps } from "../ContentContainer";

const useStyles = makeStyles((theme: Theme) => ({
  cell: {
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
                    className={clsx(record["@meta"].flash && classes.highlightedCell)}
                    hover={true}
                  >
                    {columns.map((column) => (
                      <TableCell key={column.name} className={classes.cell} align={column.isNumeric ? "right" : "left"}>
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
