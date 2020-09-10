import { Icon, makeStyles, Theme, Grid, Typography } from "@material-ui/core";
import InfoIcon from "@material-ui/icons/InfoSharp";
import React, { FC, useMemo } from "react";

import { Schema } from "../stream/schema";
import { StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName } from "apollo/types/StreamByOrganizationProjectAndName";

const useStyles = makeStyles((theme: Theme) => ({
}));

export interface SchemaViewProps {
  stream: StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName;
}

const SchemaView: FC<SchemaViewProps> = ({ stream }) => {
  const classes = useStyles();
  const schema = useMemo(() => new Schema(stream.avroSchema, stream.streamIndexes), [stream.avroSchema]);
  const columns = schema.getColumns(false);
  return (
    <Grid container direction="column">
      {columns.map((column) => (
        <Grid key={column.name} item container>
          <Grid item xs>
            <Typography>{column.name}</Typography>
          </Grid>
          <Grid item>
            <Typography>{column.type.toString()}</Typography>
          </Grid>
        </Grid>
      ))}
    </Grid>
  );
};

export default SchemaView;
