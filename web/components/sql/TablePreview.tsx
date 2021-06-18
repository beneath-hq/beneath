import { ApolloQueryResult } from "@apollo/client";
import { makeStyles, Paper, Typography, Theme, Box } from "@material-ui/core";
import _ from "lodash";
import React, { FC } from "react";

import { Link } from "components/Link";
import { TableByOrganizationProjectAndName } from "apollo/types/TableByOrganizationProjectAndName";
import { toURLName } from "lib/names";
import SchemaView from "./SchemaView";

const useStyles = makeStyles((_: Theme) => ({
  message: {
    margin: "2rem 0",
  },
  description: {
    marginTop: "0.3rem",
  },
  tablePreviewPaper: {
    padding: "12px",
  },
}));

export interface TablePreviewProps {
  result: ApolloQueryResult<TableByOrganizationProjectAndName>;
}

export const TablePreview: FC<TablePreviewProps> = ({ result: { data, error, errors } }) => {
  let errorMsg = null;
  if (error || errors || !data) {
    errorMsg = error ? error.message : errors && errors.length > 0 ? errors[0].message : "Unknown error";
  }

  const table = data?.tableByOrganizationProjectAndName;
  const orgName = toURLName(table?.project.organization.name || "");
  const projName = toURLName(table?.project.name || "");

  const classes = useStyles();
  return (
    <Paper className={classes.tablePreviewPaper}>
      {errorMsg && (
        <Typography className={classes.message} variant="h4" color="error" align="center">
          {errorMsg}
        </Typography>
      )}
      {table && (
        <>
          <Typography variant="subtitle2">Table schema</Typography>
          <Link
            color="textPrimary"
            underline="none"
            variant="h3"
            gutterBottom
            href={`/table?organization_name=${orgName}&project_name=${projName}&table_name=${table.name}`}
            as={`/${orgName}/${projName}/table:${table.name}`}
          >
            {orgName}/{projName}/{toURLName(table.name)}
          </Link>
          <Typography color="textSecondary" variant="body2" gutterBottom className={classes.description}>
            {table.description}
          </Typography>
          <Box mt={1.5}>
            <SchemaView table={table} />
          </Box>
        </>
      )}
    </Paper>
  );
};

export default TablePreview;
