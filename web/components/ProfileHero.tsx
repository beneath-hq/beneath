import { FC } from "react";
import { Grid, Link, makeStyles, Typography } from "@material-ui/core";

import Avatar from "./Avatar";

const useStyles = makeStyles((theme) => ({
  avatar: {
    marginRight: theme.spacing(3),
  },
  container: {
    marginTop: theme.spacing(4),
    marginBottom: theme.spacing(4),
  },
  site: {
    display: "block",
  },
  gutterBottom: {
    marginBottom: theme.spacing(0.75),
  },
}));

export interface ProfileHeroProps {
  name: string;
  displayName?: string | null;
  description?: string | null;
  site?: string | null;
  avatarURL?: string | null;
}

const ProfileHero: FC<ProfileHeroProps> = ({ name, displayName, description, site, avatarURL }) => {
  const classes = useStyles();
  return (
    <Grid container wrap="nowrap" spacing={0} className={classes.container}>
      <Grid className={classes.avatar} item>
        <Avatar size="hero" label={displayName || name} src={avatarURL} />
      </Grid>
      <Grid item>
        <Typography component="h1" variant="h1" className={classes.gutterBottom}>
          {displayName || name}
        </Typography>
        <Typography variant="body1" className={classes.gutterBottom}>
          {description}
        </Typography>
        {site && (
          <Link href={site} target="_blank" variant="subtitle2" className={classes.site}>
            {site}
          </Link>
        )}
      </Grid>
    </Grid>
  );
};

export default ProfileHero;
