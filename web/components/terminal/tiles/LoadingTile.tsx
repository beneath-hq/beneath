import React, { FC } from "react";

import { CircularProgress, Grid, makeStyles, Theme, Typography } from "@material-ui/core";

import { Tile, TileProps } from "./Tile";

const useStyles = makeStyles((theme: Theme) => ({
  container: {
    height: "inherit",
    minHeight: "inherit",
    padding: theme.spacing(1),
  },
}));

export const LoadingTile: FC<TileProps> = ({ ...tileProps }) => {
  const classes = useStyles();
  return (
    <Tile {...tileProps}>
      <Grid className={classes.container} container justify="center" alignContent="center" alignItems="center">
        <Grid item>
          <CircularProgress />
        </Grid>
      </Grid>
    </Tile>
  );
};

export default LoadingTile;
