import React, { FC } from "react";

import { Grid, makeStyles, Theme, Typography } from "@material-ui/core";
import { Tile, TileProps } from "./Tile";
import { ArrowRightAlt } from "@material-ui/icons";

const useStyles = makeStyles((theme: Theme) => ({
  container: {
    height: "inherit",
    minHeight: "inherit",
    padding: theme.spacing(1.5),
  },
}));

export interface ActionTileProps extends TileProps {
  title: string;
}

export const ActionTile: FC<ActionTileProps> = ({ title, children, ...tileProps }) => {
  const classes = useStyles();
  return (
    <Tile {...tileProps}>
      <Grid className={classes.container} container alignItems="center" spacing={1}>
        <Grid item>
          {children}
        </Grid>
        <Grid item>
          <Typography variant="h4">
            {title}
          </Typography>
        </Grid>
        <Grid item>
          <ArrowRightAlt />
        </Grid>
      </Grid>
    </Tile>
  );
};

export default ActionTile;
