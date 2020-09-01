import { makeStyles } from "@material-ui/core/styles";
import { FC } from "react";

import Chip from "@material-ui/core/Chip";
import Grid from "@material-ui/core/Grid";
import Typography from "@material-ui/core/Typography";

import { prettyPrintBytes, Metrics } from "./metrics/util";
import SelectField from "./SelectField";
import { NakedLink } from "./Link";
import { StreamInstancesByOrganizationProjectAndStreamName_streamInstancesByOrganizationProjectAndStreamName } from "apollo/types/StreamInstancesByOrganizationProjectAndStreamName";
import { StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName_primaryStreamInstance } from "apollo/types/StreamByOrganizationProjectAndName";

const useStyles = makeStyles((theme) => ({
  container: {
    marginTop: theme.spacing(4),
    marginBottom: theme.spacing(4),
  },
  site: {
    display: "block",
  },
}));

interface ModelHeroProps {
  name: string;
  project: string;
  organization: string;
  description: string | null;
  permissions: boolean;
  currentInstance: StreamInstancesByOrganizationProjectAndStreamName_streamInstancesByOrganizationProjectAndStreamName | StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName_primaryStreamInstance;
  instances: StreamInstancesByOrganizationProjectAndStreamName_streamInstancesByOrganizationProjectAndStreamName[];
  setInstance: (instanceID: string) => void;
  // metrics: Metrics;
}

// const ModelHero: FC<ModelHeroProps> = ({ name, project, organization, description, permissions, currentInstance, instances, setInstance, metrics }) => {
const ModelHero: FC<ModelHeroProps> = ({ name, project, organization, description, permissions, currentInstance, instances, setInstance }) => {
  // console.log(metrics)
  const classes = useStyles();
  return (
    <Grid container justify="space-between">
      <Grid item>
        <Grid container direction="column" spacing={1} className={classes.container}>
          <Grid item>
            <Grid container alignItems="center" spacing={2}>
              <Grid item>
                <Typography component="h1" variant="h1">
                  {name}
                </Typography>
              </Grid>
              <Grid item>
                <Chip
                  label={permissions ? "public" : "private"}
                  clickable
                  component={NakedLink}
                  href={`/project?organization_name=${organization}&project_name=${project}&tab=members`}
                  as={`/${organization}/${project}/-/members`}
                />
              </Grid>
            </Grid>
          </Grid>
          <Grid item>
            <Grid container spacing={1}>
              <Grid item>
                <Chip
                  label={prettyPrintBytes(23000000) + " written"}
                  clickable
                  component={NakedLink}
                  href={`/stream?organization_name=${organization}&project_name=${project}&stream_name=${name}&tab=monitoring`}
                  as={`/${organization}/${project}/${name}/-/monitoring`}
                />
              </Grid>
              <Grid item>
                <Chip
                  label={prettyPrintBytes(233333333) + " read"}
                  clickable
                  component={NakedLink}
                  href={`/stream?organization_name=${organization}&project_name=${project}&stream_name=${name}&tab=monitoring`}
                  as={`/${organization}/${project}/${name}/-/monitoring`}
                />
              </Grid>
            </Grid>
          </Grid>
          <Grid item>
            <Typography variant="body1">{description}</Typography>
          </Grid>
        </Grid>
      </Grid>
      <Grid item className={classes.container}>
        <SelectField
          id="instanceID"
          label="Instance"
          value={currentInstance.streamInstanceID}
          options={instances.map((instance) => {
            return { label: "v" + instance.version.toString(), value: instance.streamInstanceID };
          })}
          onChange={({ target }) => {
            setInstance(target.value as string);
          }}
        />
      </Grid>
    </Grid>
  );
};

export default ModelHero;
