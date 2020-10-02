import React, { FC } from "react";
import { Grid, makeStyles, Theme, Typography } from "@material-ui/core";

import { Tile, TileProps } from "./Tile";

const useStyles = makeStyles((theme: Theme) => ({
  container: {
    height: "inherit",
    minHeight: theme.spacing(15),
    padding: theme.spacing(1.5),
  },
}));

export interface PlaceholderTileProps extends TileProps {
  title: string;
}

export const PlaceholderTile: FC<PlaceholderTileProps> = ({ title, ...tileProps }) => {
  const classes = useStyles();
  return (
    <Tile {...tileProps}>
      <Grid className={classes.container} container direction="column" justify="center">
        <Grid item>
          <Typography variant="subtitle2" align="center">
            {title}
          </Typography>
        </Grid>
      </Grid>
    </Tile>
  );
};

export default PlaceholderTile;
