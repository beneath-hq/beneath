import { Chip, Grid, Typography, makeStyles, Tooltip } from "@material-ui/core";
import { FC } from "react";

import { EntityKind } from "apollo/types/globalTypes";
import { StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName } from "apollo/types/StreamByOrganizationProjectAndName";
import { NakedLink } from "components/Link";
import { StreamInstance } from "components/stream/types";
import { useTotalUsage } from "components/usage/util";
import { toURLName } from "lib/names";
import { StreamInstanceByOrganizationProjectStreamAndVersion_streamInstanceByOrganizationProjectStreamAndVersion_stream } from "apollo/types/StreamInstanceByOrganizationProjectStreamAndVersion";
import StreamInstanceSelector from "./StreamInstanceSelector";
import { makeStreamAs, makeStreamHref } from "./urls";
import { MetaChip, StreamUsageChip } from "./chips";

const useStyles = makeStyles((theme) => ({
  container: {
    marginTop: theme.spacing(2),
    marginBottom: theme.spacing(2),
  },
  streamName: {
    fontSize: theme.typography.pxToRem(30),
    fontWeight: "bold",
  },
  site: {
    display: "block",
  },
  selectField: {
    width: "170px",
  },
  dropdownButton: {
    backgroundColor: theme.palette.background.paper,
    "&:hover": {
      backgroundColor: theme.palette.secondary.main,
    },
    border: `1px solid ${theme.palette.border.paper}`,
    color: theme.palette.common.white,
  },
  verticalBar: {
    display: "inline-block",
    width: "1px",
    height: "18px",
    marginRight: "12px",
    marginLeft: "12px",
    backgroundColor: theme.palette.text.disabled,
  },
}));

export interface StreamHeroProps {
  stream: StreamInstanceByOrganizationProjectStreamAndVersion_streamInstanceByOrganizationProjectStreamAndVersion_stream;
  instance: StreamInstance | null;
}

const StreamHero: FC<StreamHeroProps> = ({ stream, instance }) => {
  const classes = useStyles();
  const organizationName = stream.project.organization.name;
  const projectName = stream.project.name;
  const streamName = stream.name;

  return (
    <Grid className={classes.container} container alignItems="center" spacing={2}>
      <Grid item>
        <Typography className={classes.streamName}>{toURLName(streamName)}</Typography>
      </Grid>
      <Grid item>
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
