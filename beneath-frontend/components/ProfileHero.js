import PropTypes from "prop-types";
import { makeStyles } from "@material-ui/core/styles";

import Avatar from "@material-ui/core/Avatar";
import Grid from "@material-ui/core/Grid";
import Link from "@material-ui/core/Link";
import Typography from "@material-ui/core/Typography";

const useStyles = makeStyles((theme) => ({
  avatar: {
    width: theme.spacing(8),
    height: theme.spacing(8),
    borderRadius: "10%",
    marginRight: theme.spacing(2),
  },
  container: {
    marginTop: theme.spacing(4),
    marginBottom: theme.spacing(4),
  },
  site: {
    display: "block",
  },
}));

const ProfileHero = ({ name, description, site, avatarUrl }) => {
  const classes = useStyles();
  return (
    <Grid container wrap="nowrap" spacing={1} className={classes.container}>
      <Grid item>
        <Avatar className={classes.avatar} src={avatarUrl} alt={name}>
        </Avatar>
      </Grid>
      <Grid item>
        <Typography component="h1" variant="h1" gutterBottom>{name}</Typography>
        {site && <Link href={site} variant="subtitle2" className={classes.site} gutterBottom>{site}</Link>}
        <Typography variant="body1">{description}</Typography>
      </Grid>
    </Grid>
  );
};

ProfileHero.propTypes = {
  name: PropTypes.string,
  description: PropTypes.string,
  site: PropTypes.string,
  avatarUrl: PropTypes.string,
};

export default ProfileHero;