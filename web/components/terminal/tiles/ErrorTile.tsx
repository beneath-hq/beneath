import React, { FC } from "react";

import { Grid, makeStyles, Theme, Typography } from "@material-ui/core";

import { Tile, TileProps } from "./Tile";

const useStyles = makeStyles((theme: Theme) => ({
  container: {
    height: "inherit",
    minHeight: "inherit",
    padding: theme.spacing(2),
  },
}));

export interface ErrorTileProps extends TileProps {
  error: string;
}

export const ErrorTile: FC<ErrorTileProps> = ({ error, ...tileProps }) => {
  const classes = useStyles();
  return (
    <Tile {...tileProps}>
      <Grid className={classes.container} container justify="center" alignContent="center" alignItems="center">
        <Grid item>
          <Typography variant="body1" color="error" align="center">
            {error}
          </Typography>
        </Grid>
      </Grid>
    </Tile>
  );
};

export default ErrorTile;
