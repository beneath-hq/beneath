import Button from "@material-ui/core/Button";
import Grid from "@material-ui/core/Grid";
import Typography from "@material-ui/core/Typography";
import { withStyles } from '@material-ui/core/styles';

import connection from "../lib/connection";
import Page from "../components/Page";
import { GoogleIcon, GithubIcon } from "../components/Icons";

const styles = (theme) => ({
  authContainer: {
    maxWidth: 900,
    margin: '0 auto',
    paddingTop: theme.spacing.unit * 8,
  },
  authButtonsContainer: {
    paddingTop: theme.spacing.unit * 4,
  },
  icon: {
    fontSize: 24,
    marginRight: theme.spacing.unit,
  },
});

export default withStyles(styles)(({ classes }) => (
  <Page title="Sign Up or Log In">
    <div className={ classes.authContainer }>
      <Typography component="h2" variant="h5" align="center" gutterBottom>
        Sign Up or Log In
      </Typography>
      <div className={classes.authButtonsContainer}>
        <Grid container spacing={16} justify="center">
          <Grid item>
            <Button size="large" color="primary" variant="outlined" href={`${connection.API_URL}/auth/github`}>
              <GithubIcon className={classes.icon} />
              Connect with Github
            </Button>
          </Grid>
          <Grid item>
            <Button size="large" color="primary" variant="outlined" href={`${connection.API_URL}/auth/google`}>
              <GoogleIcon className={classes.icon} />
              Connect with Google
            </Button>
          </Grid>
        </Grid>
      </div>
    </div>
  </Page>
));
