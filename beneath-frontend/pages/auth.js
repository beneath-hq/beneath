import Button from "@material-ui/core/Button";
import Container from "@material-ui/core/Container";
import Grid from "@material-ui/core/Grid";
import Typography from "@material-ui/core/Typography";
import { withStyles } from "@material-ui/styles";

import connection from "../lib/connection";
import Page from "../components/Page";
import { GoogleIcon, GithubIcon } from "../components/Icons";

const styles = (theme) => ({
  authContent: {
    padding: theme.spacing(8, 0, 6),
  },
  authButtons: {
    marginTop: theme.spacing(4),
  },
  icon: {
    fontSize: 24,
    marginRight: theme.spacing(1),
  },
});

export default withStyles(styles)(({ classes }) => (
  <Page title="Sign Up or Log In">
    <div className={classes.authContent}>
      <Container maxWidth="lg">
        <Typography component="h2" variant="h5" align="center" gutterBottom>
          Sign Up or Log In
        </Typography>
        <div className={classes.authButtons}>
          <Grid container spacing={2} justify="center">
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
      </Container>
    </div>
  </Page>
));
