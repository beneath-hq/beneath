import React, { FC } from "react";

import { Grid, makeStyles, Theme, Typography } from "@material-ui/core";

import { Tile, TileProps } from "./Tile";

const useStyles = makeStyles((theme: Theme) => ({
  container: {
    height: "inherit",
    minHeight: "inherit",
    padding: theme.spacing(2),
  },
  title: {
    fontSize: theme.typography.pxToRem(28),
  },
}));

export interface ActionTileProps extends TileProps {
  title: string;
}

export const ActionTile: FC<ActionTileProps> = ({ title, ...tileProps }) => {
  const classes = useStyles();
  return (
    <Tile {...tileProps}>
      <Grid className={classes.container} container justify="center" alignContent="center" alignItems="center">
        <Grid item>
          <Typography className={classes.title} component="h2" variant="h2" color="primary" align="center">
            {title}
          </Typography>
        </Grid>
      </Grid>
    </Tile>
  );
};

export default ActionTile;
