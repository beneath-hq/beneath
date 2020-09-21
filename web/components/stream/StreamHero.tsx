import { useQuery } from "@apollo/client";
import { Chip, Grid, Typography, makeStyles } from "@material-ui/core";
import { FC } from "react";

import { NakedLink } from "../Link";
import { useMonthlyMetrics } from "../metrics/hooks";
import { prettyPrintBytes } from "../metrics/util";
import SelectField from "../SelectField";
import { EntityKind } from "apollo/types/globalTypes";
import { QUERY_STREAM_INSTANCES } from "apollo/queries/stream";
import {
  StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName_primaryStreamInstance,
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

export interface StreamHeroProps {
  stream: StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName;
  instance: StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName_primaryStreamInstance | null;
  setInstance: (
    instance: StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName_primaryStreamInstance | null
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

  let instanceOptions: StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName_primaryStreamInstance[] = [];
  if (data) {
    instanceOptions = data.streamInstancesByOrganizationProjectAndStreamName;
  } else if (stream.primaryStreamInstance) {
    instanceOptions.push(stream.primaryStreamInstance);
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
          value={instance?.streamInstanceID || null}
          options={
            !data
              ? []
              : data.streamInstancesByOrganizationProjectAndStreamName.map((instance) => {
                  return { label: "v" + instance.version.toString(), value: instance.streamInstanceID };
                })
          }
          onChange={({ target }) => {
            const instanceID = target.value;
            const newInstance = instanceOptions.find((instance) => instance.streamInstanceID === instanceID);
            setInstance(newInstance || null);
          }}
        />
      </Grid>
    </Grid>
  );
};

export default StreamHero;
