import React, { FC } from "react";
import { Button, Grid } from "@material-ui/core";

import { Tile, TileProps } from "./Tile";
import useMe from "hooks/useMe";

export const ActionsTile: FC<TileProps> = ({ ...tileProps }) => {
  const me = useMe();

  if (!me) {
    return <></>;
  }

  return (
    <Tile {...tileProps}>
      <Grid container alignItems="center" spacing={2}>
        <Grid item>
          <Button
            variant="contained"
            color="secondary"
            href="https://about.beneath.dev/docs/quick-starts/"
            target="_blank"
            size="small"
          >
            Quick starts
          </Button>
        </Grid>
        <Grid item>
          <Button
            variant="contained"
            color="secondary"
            href="https://about.beneath.dev/contact"
            target="_blank"
            size="small"
          >
            Contact support
          </Button>
        </Grid>
        <Grid item>
          <Button
            variant="contained"
            color="secondary"
            size="small"
            href="https://discord.gg/f5yvx7YWau"
            target="_blank"
          >
            Join the Discord community
          </Button>
        </Grid>
      </Grid>
    </Tile>
  );
};

export default ActionsTile;
