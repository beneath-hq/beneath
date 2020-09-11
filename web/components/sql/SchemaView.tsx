import { Chip, makeStyles, Theme, Grid, Tooltip, Typography } from "@material-ui/core";
import React, { FC, useMemo } from "react";

import { Schema } from "../stream/schema";
import { StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName } from "apollo/types/StreamByOrganizationProjectAndName";

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
  stream: StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName;
}

const SchemaView: FC<SchemaViewProps> = ({ stream }) => {
  const classes = useStyles();
  const schema = useMemo(() => new Schema(stream.avroSchema, stream.streamIndexes), [stream.avroSchema]);
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
