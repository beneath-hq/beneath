import { Grid, makeStyles, Theme, Typography } from "@material-ui/core";
import React, { FC } from "react";

import ExploreProjectsTiles from "./ExploreProjectsTiles";
import ActionsTile from "./tiles/ActionsTile";

const useStyles = makeStyles((theme: Theme) => ({
  sectionTitle: {
    marginTop: theme.spacing(4),
  },
  title: {
    fontSize: theme.typography.pxToRem(32),
    marginBottom: theme.spacing(2)
  },
}));

const Welcome: FC = () => {
  const classes = useStyles();

  return (
    <Grid container spacing={3}>
      <Grid item xs={12}>
        <Typography variant="h1" className={classes.title}>
          Welcome to Beneath!
        </Typography>
      </Grid>

      <ActionsTile shape="wide" nopaper />

      <Grid item xs={12}>
        <Typography variant="h3" className={classes.sectionTitle}>
          Featured projects and tutorials
        </Typography>
      </Grid>
      <ExploreProjectsTiles />
    </Grid>
  );
};

export default Welcome;
