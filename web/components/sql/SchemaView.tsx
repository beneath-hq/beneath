import { Chip, makeStyles, Theme, Grid, Tooltip, Typography } from "@material-ui/core";
import React, { FC, useMemo } from "react";

import { Schema } from "../table/schema";
import { TableByOrganizationProjectAndName_tableByOrganizationProjectAndName } from "apollo/types/TableByOrganizationProjectAndName";

const useStyles = makeStyles((theme: Theme) => ({
  icon: {
    fontSize: theme.typography.body1.fontSize,
  },
  fieldName: {
    fontFamily: theme.typography.fontFamilyMonospaced,
    fontSize: theme.typography.body2.fontSize,
  },
}));

export interface SchemaViewProps {
  table: TableByOrganizationProjectAndName_tableByOrganizationProjectAndName;
}

const SchemaView: FC<SchemaViewProps> = ({ table }) => {
  const classes = useStyles();
  const schema = useMemo(() => new Schema(table.avroSchema, table.tableIndexes), [table.avroSchema]);
  const columns = schema.getColumns(false);
  return (
    <Grid container direction="column" spacing={1}>
      {columns.map((column) => (
        <Grid key={column.name} item>
          <Grid container alignItems="center" spacing={1}>
            {column.doc && (
              <Grid item xs>
                <Tooltip title={column.doc} placement="bottom-start">
                  <Typography className={classes.fieldName}>{column.name}</Typography>
                </Tooltip>
              </Grid>
            )}
            {!column.doc && (
              <Grid item xs>
                <Typography className={classes.fieldName}>{column.name}</Typography>
              </Grid>
            )}
            {column.isNullable && (
              <Grid item>
                <Chip variant="outlined" size="small" label="Optional" color="secondary" />
              </Grid>
            )}
            <Grid item>
              <Tooltip title={column.typeDescription}>
                <Chip variant="outlined" size="small" label={column.typeName} />
              </Tooltip>
            </Grid>
          </Grid>
        </Grid>
      ))}
    </Grid>
  );
};

export default SchemaView;
