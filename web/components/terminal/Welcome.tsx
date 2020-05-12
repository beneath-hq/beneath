import { Grid, makeStyles, Theme } from "@material-ui/core";
import React, { FC } from "react";

import ExploreProjectsTiles from "./ExploreProjectsTiles";
import ActionTile from "./tiles/ActionTile";
import TitleTile from "./tiles/TitleTile";

const useStyles = makeStyles((theme: Theme) => ({}));

const Welcome: FC = () => {
  const classes = useStyles();

  return (
    <Grid container spacing={2} alignItems="stretch" alignContent="stretch">
      <TitleTile title="Welcome to Beneath!" />
      <ActionTile title="Create Account" href={`/-/auth`} as={`/-/auth`} />
      <ActionTile title="Tutorials" href="https://about.beneath.dev/docs/quick-starts/" />
      <ActionTile title="Contact us" href="https://about.beneath.dev/contact/" />
      <TitleTile title="Featured projects:" />
      <ExploreProjectsTiles />
    </Grid>
  );
};

export default Welcome;
