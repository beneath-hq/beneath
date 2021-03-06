import { Record } from "beneath-react";
import clsx from "clsx";
import React, { FC } from "react";

import {
  Button,
  Chip,
  makeStyles,
  Table as MuiTable,
  TableBody as MuiTableBody,
  TableCell as MuiTableCell,
  TableHead as MuiTableHead,
  TableRow as MuiTableRow,
  Theme,
  Tooltip,
  Box,
  Typography,
  fade,
  Grid,
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
  headerCellText: {
    marginRight: theme.spacing(1),
    fontSize: theme.typography.pxToRem(16),
    fontWeight: "bold",
  },
  headerCellInfo: {
    fontSize: theme.typography.h3.fontSize,
    marginTop: theme.spacing(0.25), // total hack to center the InfoIcon in the Chip
  },
  primaryDarkColor: {
    backgroundColor: theme.palette.primary.dark,
  },
}));

export interface RecordsViewProps extends ContentContainerProps {
  schema?: Schema;
  records?: Record[];
  showTimestamps?: boolean;
  fetchMore?: () => void;
}

const RecordsView: FC<RecordsViewProps> = ({
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
      <ContentContainer loading={loading} {...containerProps}>
        {columns && (
          <MuiTable size="small">
            <MuiTableHead>
              <MuiTableRow>
                {columns.map((column) => (
                  <MuiTableCell key={column.name} className={clsx(classes.cell, classes.headerCell)}>
                    <Grid container spacing={1} alignItems="center" wrap="nowrap">
                      <Grid item>
                        <Typography className={classes.headerCellText}>{column.displayName}</Typography>
                      </Grid>
                      {column.isKey && (
                        <Grid item>
                          <Tooltip title={"Column is part of an index"} interactive>
                            <Chip label="Key" size="small" className={classes.primaryDarkColor} />
                          </Tooltip>
                        </Grid>
                      )}
                      {column.displayName !== "Time ago" && (
                        <Grid item>
                          <Tooltip title={column.typeDescription} interactive>
                            <Chip label={column.typeName} size="small" />
                          </Tooltip>
                        </Grid>
                      )}
                      {column.doc && (
                        <Grid item>
                          <Tooltip title={column.doc} interactive>
                            <Chip label={<InfoIcon className={classes.headerCellInfo} />} size="small" />
                          </Tooltip>
                        </Grid>
                      )}
                    </Grid>
                  </MuiTableCell>
                ))}
              </MuiTableRow>
            </MuiTableHead>
            <MuiTableBody>
              {records &&
                records.map((record, _) => (
                  <MuiTableRow
                    key={`${record["@meta"].key};${record["@meta"].timestamp}`}
                    className={clsx(record["@meta"].flash && classes.highlightedCell)}
                    hover={true}
                  >
                    {columns.map((column) => (
                      <MuiTableCell
                        key={column.name}
                        className={clsx(classes.cell, column.isKey && classes.keyCell)}
                        align={column.isNumeric ? "right" : "left"}
                      >
                        {column.formatRecord(record)}
                      </MuiTableCell>
                    ))}
                  </MuiTableRow>
                ))}
            </MuiTableBody>
          </MuiTable>
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

export default RecordsView;
