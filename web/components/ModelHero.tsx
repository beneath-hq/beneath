import PropTypes from "prop-types";
import { makeStyles } from "@material-ui/core/styles";
import { FC } from "react";

import Chip from "@material-ui/core/Chip";
import Grid from "@material-ui/core/Grid";
import Typography from "@material-ui/core/Typography";
import { prettyPrintBytes, Metrics } from "./metrics/util";
import SelectField from "./SelectField";

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
  description: string | null;
  permissions: boolean;
  // metrics: Metrics;
}

// const ModelHero: FC<ModelHeroProps> = ({ name, description, permissions, metrics }) => {
const ModelHero: FC<ModelHeroProps> = ({ name, description, permissions }) => {
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
                <Chip label={permissions ? "public" : "private"}/>
              </Grid>
            </Grid>
          </Grid>
          <Grid item>
            <Grid container spacing={1}>
              <Grid item>
                <Chip label={prettyPrintBytes(23000000) + " written"} />
              </Grid>
              <Grid item>
                <Chip label={prettyPrintBytes(233333333) + " read"} />
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
          id="instance"
          label="Instance"
          value={"Instance2"}
          options={[
            { label: "Instance1", value: "Instance1" },
            { label: "Instance2", value: "Instance2" },
          ]}
          // onChange={({ target }) => setInstance(target.value)}
          // controlClass={classes.selectPeekControl}
        />
      </Grid>
    </Grid>
  );
};

export default ModelHero;
