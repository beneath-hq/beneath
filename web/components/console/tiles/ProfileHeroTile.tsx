import { FC } from "react";
import { Grid, makeStyles, Typography } from "@material-ui/core";

import Avatar from "../../Avatar";
import { Tile, TileProps } from "./Tile";

const useStyles = makeStyles((theme) => ({
  avatar: {
    paddingRight: theme.spacing(3),
  },
  container: {
    padding: theme.spacing(2),
  },
  description: {
    marginTop: theme.spacing(3),
  },
}));

export interface ProfileHeroTileProps extends TileProps {
  name: string;
  displayName?: string | null;
  path?: string | null;
  description?: string | null;
  avatarURL?: string | null;
}

const ProfileHeroTile: FC<ProfileHeroTileProps> = ({ name, displayName, path, description, avatarURL, shape, ...tileProps }) => {
  const classes = useStyles();
  return (
    <Tile shape={shape} {...tileProps}>
      <Grid container spacing={0} className={classes.container} alignItems="center">
        <Grid className={classes.avatar} item>
          <Avatar size="hero" label={displayName || name} src={avatarURL} />
        </Grid>
        <Grid item>
          <Typography variant="h3">{displayName || name}</Typography>
          <Typography variant="subtitle1" color="textSecondary">
            {path}
          </Typography>
        </Grid>
        {description && (
          <Grid item xs={12}>
            <Typography className={classes.description} variant="body1">
              {description}
            </Typography>
          </Grid>
        )}
      </Grid>
    </Tile>
  );
};

export default ProfileHeroTile;
