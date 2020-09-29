import { Grid, makeStyles, Theme, Typography } from "@material-ui/core";
import { ArrowRightAlt, Folder, LinearScale, VpnKey } from "@material-ui/icons";
import React, { FC } from "react";

import { Link } from "components/Link";
import ExploreProjectsTiles from "./ExploreProjectsTiles";
import ActionTile from "./tiles/ActionTile";

const useStyles = makeStyles((theme: Theme) => ({
  sectionTitle: {
    marginTop: theme.spacing(4),
  },
  title: {
    fontSize: theme.typography.pxToRem(32),
  },
}));

const Welcome: FC = () => {
  const classes = useStyles();

  return (
    <Grid container spacing={3}>
      <Grid item xs={12}>
        <Typography variant="h1" className={classes.title}>Welcome to Beneath!</Typography>
      </Grid>

      <Grid item xs={12}>
        <Typography variant="h3" className={classes.sectionTitle}>
          Create new
        </Typography>
      </Grid>
      <ActionTile title="Project" href={`/-/auth`} shape="dense">
        <Folder color="primary" />
      </ActionTile>
      <ActionTile title="Stream" href={`/-/auth`} shape="dense">
        <LinearScale color="primary" />
      </ActionTile>
      <ActionTile
        title="Secret"
        href={`/-/auth`}
        shape="dense"
      >
        <VpnKey color="primary" />
      </ActionTile>

      <Grid item xs={12}>
        <Link href="https://about.beneath.dev/docs/quick-starts/">
          <Grid container spacing={1} alignItems="center">
            <Grid item>
              <Typography variant="h3">Read the quick start</Typography>
            </Grid>
            <Grid item>
              <ArrowRightAlt />
            </Grid>
          </Grid>
        </Link>
      </Grid>

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
