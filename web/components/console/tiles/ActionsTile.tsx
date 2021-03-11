import React, { FC } from "react";
import { Button, Grid, makeStyles, Theme } from "@material-ui/core";

import { Tile, TileProps } from "./Tile";
import useMe from "hooks/useMe";

const useStyles = makeStyles((theme: Theme) => ({
  rightButton: {
    height: "42px",
  },
  icon: {
    fill: theme.palette.primary.dark,
  },
}));

export const ActionsTile: FC<TileProps> = ({ ...tileProps }) => {
  const classes = useStyles();
  const me = useMe();

  if (!me) {
    return <></>;
  }

  return (
    <Tile {...tileProps}>
      <Grid container alignItems="center" spacing={2}>
        <Grid item>
          <Button variant="contained" color="secondary" href="https://about.beneath.dev/contact" size="small">
            Contact support
          </Button>
        </Grid>
        <Grid item>
          <Button variant="contained" color="secondary" size="small" href="https://discord.gg/f5yvx7YWau">
            Join the Discord community
          </Button>
        </Grid>
      </Grid>
    </Tile>
  );
};

export default ActionsTile;
