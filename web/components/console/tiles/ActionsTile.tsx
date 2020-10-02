import React, { FC } from "react";
import { Button, Grid, makeStyles, Theme } from "@material-ui/core";
import { Code, Folder, Functions, LinearScale, Mail, MenuBook, VpnKey } from "@material-ui/icons";

import { Tile, TileProps } from "./Tile";
import useMe from "hooks/useMe";

const useStyles = makeStyles((theme: Theme) => ({
  rightButton: {
    height: "42px",
  },
  icon: {
    fill: theme.palette.primary.dark
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
          <Button
            variant="contained"
            color="secondary"
            size="small"
            startIcon={<Folder className={classes.icon} />}
            href="/-/create/project"
          >
            Create project
          </Button>
        </Grid>
        <Grid item>
          <Button
            variant="contained"
            color="secondary"
            size="small"
            startIcon={<LinearScale className={classes.icon} />}
            href="/-/create/stream"
          >
            Create stream
          </Button>
        </Grid>
        <Grid item>
          <Button variant="contained" color="secondary" size="small" startIcon={<Functions className={classes.icon} />}>
            Create service
          </Button>
        </Grid>
        <Grid item>
          <Button
            variant="contained"
            color="secondary"
            size="small"
            startIcon={<VpnKey className={classes.icon} />}
            href={`/organization?organization_name=${me.name}&tab=secrets`}
          >
            Create secret
          </Button>
        </Grid>
        <Grid item>
          <Button
            variant="contained"
            color="secondary"
            size="small"
            startIcon={<Code className={classes.icon} />}
            href="/-/sql"
          >
            SQL editor
          </Button>
        </Grid>
        <Grid item>
          <Button
            variant="contained"
            color="secondary"
            size="small"
            startIcon={<MenuBook className={classes.icon} />}
            href="https://about.beneath.dev/docs/quick-starts/"
          >
            Quick start
          </Button>
        </Grid>
        <Grid item>
          <Button
            variant="contained"
            color="secondary"
            size="small"
            startIcon={<Mail className={classes.icon} />}
            href="https://about.beneath.dev/contact"
          >
            Contact us
          </Button>
        </Grid>
      </Grid>
    </Tile>
  );
};

export default ActionsTile;
