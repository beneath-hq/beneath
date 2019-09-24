import PropTypes from "prop-types";
import { makeStyles } from "@material-ui/core/styles";

import Grid from "@material-ui/core/Grid";
import Link from "@material-ui/core/Link";
import Typography from "@material-ui/core/Typography";

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
}));

const ProfileHero = ({ name, description, site, avatarURL }) => {
  const classes = useStyles();
  return (
    <Grid container wrap="nowrap" spacing={0} className={classes.container}>
      <Grid className={classes.avatar} item>
        <Avatar size="hero" label={name} src={avatarURL} />
      </Grid>
      <Grid item>
        <Typography component="h1" variant="h1" gutterBottom={!site}>
          {name}
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

ProfileHero.propTypes = {
  name: PropTypes.string,
  description: PropTypes.string,
  site: PropTypes.string,
  avatarURL: PropTypes.string,
};

export default ProfileHero;