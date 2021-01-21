import { Chip, Grid, Typography, makeStyles } from "@material-ui/core";
import { FC } from "react";
import numbro from "numbro";

import { EntityKind } from "apollo/types/globalTypes";
import { StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName } from "apollo/types/StreamByOrganizationProjectAndName";
import { NakedLink } from "components/Link";
import { StreamInstance } from "components/stream/types";
import { useTotalUsage } from "components/usage/util";
import { toURLName } from "lib/names";
import { StreamInstanceByOrganizationProjectStreamAndVersion_streamInstanceByOrganizationProjectStreamAndVersion_stream } from "apollo/types/StreamInstanceByOrganizationProjectStreamAndVersion";
import StreamInstanceSelector from "./StreamInstanceSelector";

const intFormat = { thousandSeparated: true };
const bytesFormat: numbro.Format = { base: "decimal", mantissa: 1, optionalMantissa: true, output: "byte" };

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
        <Grid container spacing={1}>
          <Grid item>
            <Chip
              label={stream.project.public ? "Public" : "Private"}
              clickable
              component={NakedLink}
              href={`/project?organization_name=${organizationName}&project_name=${projectName}&tab=members`}
              as={`/${organizationName}/${projectName}/-/members`}
            />
          </Grid>
          {instance && <InstanceUsageChips stream={stream} instance={instance} />}
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

interface InstanceUsageChips {
  stream: StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName;
  instance: StreamInstance;
}

// separate component to enable nested useTotalUsage
const InstanceUsageChips: FC<InstanceUsageChips> = ({ stream, instance }) => {
  const { data, loading, error } = useTotalUsage(EntityKind.StreamInstance, instance.streamInstanceID);
  if (!data) {
    return null;
  }

  const organizationName = stream.project.organization.name;
  const projectName = stream.project.name;
  const streamName = stream.name;
  const href = `/stream?organization_name=${organizationName}&project_name=${projectName}&stream_name=${streamName}&tab=monitoring`;
  const as = `/${organizationName}/${projectName}/stream:${streamName}/-/monitoring`;

  return (
    <>
      <Grid item>
        <Chip
          label={numbro(data.writeRecords).format(intFormat) + " records"}
          clickable
          component={NakedLink}
          href={href}
          as={as}
        />
      </Grid>
      <Grid item>
        <Chip label={numbro(data.writeBytes).format(bytesFormat)} clickable component={NakedLink} href={href} as={as} />
      </Grid>
    </>
  );
};
