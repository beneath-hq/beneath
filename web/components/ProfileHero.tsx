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
    marginTop: theme.spacing(0.5),
    display: "block",
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
        <Typography component="h1" variant="h1" gutterBottom={!site}>
          {displayName || name}
        </Typography>
        {site && (
          <Link href={site} variant="subtitle2" className={classes.site} gutterBottom>
            {site}
          </Link>
        )}
        <Typography variant="body1">{description}</Typography>
      </Grid>
    </Grid>
  );
};

export default ProfileHero;
