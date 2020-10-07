import { useQuery } from "@apollo/client";
import { Chip, Grid, Typography, makeStyles } from "@material-ui/core";
import { FC } from "react";

import { NakedLink } from "../Link";
import { useMonthlyMetrics } from "../metrics/hooks";
import { prettyPrintBytes } from "../metrics/util";
import SelectField from "components/forms/SelectField";
import { EntityKind } from "apollo/types/globalTypes";
import { QUERY_STREAM_INSTANCES } from "apollo/queries/stream";
import {
  StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName,
} from "apollo/types/StreamByOrganizationProjectAndName";
import {
  StreamInstancesByOrganizationProjectAndStreamNameVariables,
  StreamInstancesByOrganizationProjectAndStreamName,
} from "apollo/types/StreamInstancesByOrganizationProjectAndStreamName";

const useStyles = makeStyles((theme) => ({
  container: {
    marginTop: theme.spacing(2),
    marginBottom: theme.spacing(2),
  },
  site: {
    display: "block",
  },
}));

interface Instance {
  streamInstanceID: string;
  version: number;
}

export interface StreamHeroProps {
  stream: StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName;
  instance: Instance | undefined;
  setInstance: (
    instance: Instance | null
  ) => void;
}

const StreamHero: FC<StreamHeroProps> = ({ stream, instance, setInstance }) => {
  const organizationName = stream.project.organization.name;
  const projectName = stream.project.name;
  const streamName = stream.name;

  const metrics = useMonthlyMetrics(EntityKind.Stream, stream.streamID).total;

  const { error, data } = useQuery<
    StreamInstancesByOrganizationProjectAndStreamName,
    StreamInstancesByOrganizationProjectAndStreamNameVariables
  >(QUERY_STREAM_INSTANCES, {
    variables: { organizationName, projectName, streamName },
  });
  if (error) {
    console.error("Unexpected error loading instances: ", error);
  }

  const instances: Instance[] = [];
  if (data) {
    instances.push(...data.streamInstancesByOrganizationProjectAndStreamName);
  } else if (stream.primaryStreamInstance) {
    instances.push(stream.primaryStreamInstance);
  }

  const classes = useStyles();
  return (
    <Grid container justify="space-between" alignItems="flex-start" spacing={4} className={classes.container}>
      <Grid item>
        <Grid container direction="column" spacing={1}>
          <Grid item>
            <Grid container alignItems="center" spacing={2}>
              <Grid item>
                <Typography component="h1" variant="h1">
                  {streamName}
                </Typography>
              </Grid>
              <Grid item>
                <Chip
                  label={stream.project.public ? "public" : "private"}
                  clickable
                  component={NakedLink}
                  href={`/project?organization_name=${organizationName}&project_name=${projectName}&tab=members`}
                  as={`/${organizationName}/${projectName}/-/members`}
                />
              </Grid>
              <Grid item>
                <Chip
                  label={prettyPrintBytes(metrics.writeBytes) + " written"}
                  clickable
                  component={NakedLink}
                  href={`/stream?organization_name=${organizationName}&project_name=${projectName}&stream_name=${streamName}&tab=monitoring`}
                  as={`/${organizationName}/${projectName}/${streamName}/-/monitoring`}
                />
              </Grid>
              <Grid item>
                <Chip
                  label={prettyPrintBytes(metrics.readBytes) + " read"}
                  clickable
                  component={NakedLink}
                  href={`/stream?organization_name=${organizationName}&project_name=${projectName}&stream_name=${streamName}&tab=monitoring`}
                  as={`/${organizationName}/${projectName}/${streamName}/-/monitoring`}
                />
              </Grid>
            </Grid>
          </Grid>
          <Grid item>
            <Typography variant="body1">{stream.description}</Typography>
          </Grid>
        </Grid>
      </Grid>
      <Grid item>
        <SelectField
          id="instanceID"
          label="Instance"
          required
          options={instances}
          getOptionLabel={(option: Instance) => `v${option.version.toString()}`}
          getOptionSelected={(option: Instance, value: Instance) => {
            return option.version === value.version;
          }}
          value={instance}
          multiple={false}
          onChange={( _, value ) => {
            if (value) {
              setInstance(value as Instance);
            }
          }}
          margin="none"
        />
      </Grid>
    </Grid>
  );
};

export default StreamHero;
