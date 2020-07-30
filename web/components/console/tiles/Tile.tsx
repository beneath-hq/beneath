import clsx from "clsx";
import Link from "next/link";
import React, { FC } from "react";

import { Grid, makeStyles, Paper, Theme } from "@material-ui/core";

const useStyles = makeStyles((theme: Theme) => ({
  paper: {
    height: "100%",
    minHeight: "125px",
  },
  paperLink: {
    cursor: "pointer",
  },
  unstyledA: {
    color: "inherit",
    textDecoration: "inherit",
  }
}));

// tslint:disable-next-line: no-empty-interface
export interface TileProps {
  shape?: "normal" | "wide";
  href?: string;
  as?: string;
  nopaper?: boolean;
}

export const Tile: FC<TileProps> = ({ shape, href, as, nopaper, children }) => {
  if (!shape) {
    shape = "normal";
  }

  const classes = useStyles();

  let tile;
  if (nopaper) {
    tile = <div className={clsx(classes.paper, href && classes.paperLink)}>{children}</div>;
  } else {
    tile = <Paper className={clsx(classes.paper, href && classes.paperLink)}>{children}</Paper>;
  }

  if (href) {
    if (as) {
      tile = <Link href={href} as={as}>{tile}</Link>;
    } else {
      tile = <a className={classes.unstyledA} href={href}>{tile}</a>;
    }
  }
  return (
    <Grid item xs={12} sm={shape === "wide" ? 12 : 6} md={shape === "wide" ? 8 : 4} lg={shape === "wide" ? 6 : 3}>
      {tile}
    </Grid>
  );
};

export default Tile;
