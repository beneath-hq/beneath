import { Chip, Grid, Typography, makeStyles, Tooltip } from "@material-ui/core";
import { FC } from "react";

import { StreamInstanceByOrganizationProjectStreamAndVersion_streamInstanceByOrganizationProjectStreamAndVersion_stream } from "apollo/types/StreamInstanceByOrganizationProjectStreamAndVersion";
import { StreamInstance } from "components/stream/types";
import { toURLName } from "lib/names";
import StreamInstanceSelector from "./StreamInstanceSelector";
import { MetaChip, StreamUsageChip } from "./chips";

const useStyles = makeStyles((theme) => ({
  heroContainer: {
    marginTop: theme.spacing(2),
    marginBottom: theme.spacing(2),
  },
  streamName: {
    fontSize: theme.typography.pxToRem(30),
    fontWeight: "bold",
  },
  chipContainer: {
    [theme.breakpoints.up("md")]: {
      marginLeft: theme.spacing(1),
    },
  },
}));

export interface StreamHeroProps {
  stream: StreamInstanceByOrganizationProjectStreamAndVersion_streamInstanceByOrganizationProjectStreamAndVersion_stream;
  instance: StreamInstance | null;
}

const StreamHero: FC<StreamHeroProps> = ({ stream, instance }) => {
  const classes = useStyles();

  return (
    <Grid className={classes.heroContainer} container alignItems="center" spacing={1}>
      <Grid item>
        <Typography className={classes.streamName}>{toURLName(stream.name)}</Typography>
      </Grid>
      <Grid item className={classes.chipContainer}>
        <Grid container spacing={1} wrap="nowrap">
          {stream.meta && (
            <Grid item>
              <MetaChip />
            </Grid>
          )}
          <Grid item>
            <Tooltip
              title={
                stream.project.public
                  ? "The stream belongs to a public project and can be accessed by anyone without permission"
                  : "The stream belongs to a private project and cannot be accessed without permission"
              }
            >
              <Chip label={stream.project.public ? "Public" : "Private"} />
            </Tooltip>
          </Grid>
          {instance && (
            <Grid item>
              <StreamUsageChip stream={stream} instance={instance} />
            </Grid>
          )}
        </Grid>
      </Grid>
      <Grid item sm />
      <Grid item>
        <StreamInstanceSelector stream={stream} currentInstance={instance} />
      </Grid>
      <Grid item xs={12}>
        <Typography variant="body1">{stream.description}</Typography>
      </Grid>
    </Grid>
  );
};

export default StreamHero;
