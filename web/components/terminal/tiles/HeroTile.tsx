import { FC } from "react";

import { Grid, Link, makeStyles, Typography } from "@material-ui/core";

import Avatar from "../../Avatar";
import { Tile, TileProps } from "./Tile";

const useStyles = makeStyles((theme) => ({
  avatar: {
    paddingRight: theme.spacing(3),
  },
  container: {
    padding: theme.spacing(4),
  },
  displayName: {
    marginBottom: theme.spacing(0.75),
  },
  path: {
    marginBottom: theme.spacing(0.75),
  },
  descriptionNotWide: {
    marginTop: theme.spacing(0.75),
  },
}));

export interface HeroTileProps extends TileProps {
  name: string;
  displayName?: string | null;
  path?: string | null;
  description?: string | null;
  avatarURL?: string | null;
}

const HeroTile: FC<HeroTileProps> = ({ name, displayName, path, description, avatarURL, shape, ...tileProps }) => {
  const classes = useStyles();
  return (
    <Tile shape={shape} {...tileProps}>
      <Grid container spacing={0} className={classes.container}>
        <Grid className={classes.avatar} item>
          <Avatar size="hero" label={displayName || name} src={avatarURL} />
        </Grid>
        <Grid item>
          <Typography className={classes.displayName} component="h2" variant="h2">
            {displayName || name}
          </Typography>
          <Typography className={classes.path} variant="subtitle1" color="textSecondary">
            {path}
          </Typography>
          {shape === "wide" && <Typography variant="body1">{description}</Typography>}
        </Grid>
        {(!shape || shape !== "wide") && (
          <Grid item xs={12}>
            <Typography className={classes.descriptionNotWide} variant="body1">
              {description}
            </Typography>
          </Grid>
        )}
      </Grid>
    </Tile>
  );
};

export default HeroTile;
