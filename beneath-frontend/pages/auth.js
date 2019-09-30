import Button from "@material-ui/core/Button";
import Container from "@material-ui/core/Container";
import Grid from "@material-ui/core/Grid";
import Typography from "@material-ui/core/Typography";
import { withStyles } from "@material-ui/styles";

import connection from "../lib/connection";
import Page from "../components/Page";
import { GoogleIcon, GithubIcon } from "../components/Icons";
import VSpace from "../components/VSpace";

const styles = (theme) => ({
  authButtons: {
    marginTop: theme.spacing(4),
  },
  authButton: {
    width: "100%",
  },
  icon: {
    fontSize: 24,
    marginRight: theme.spacing(1),
  },
  title: {
    lineHeight: "150%",
    marginBottom: theme.spacing(2),
  },
});

export default withStyles(styles)(({ classes }) => (
  <Page title="Register or Login" contentMarginTop="normal">
    <div>
      <Container maxWidth="xs">
        <Typography className={classes.title} component="h2" variant="h1" align="center">
          Hello there! Pick an option to register or login
        </Typography>
        <div className={classes.authButtons}>
          <Grid container spacing={2} justify="center">
            <Grid item xs={12}>
              <Button
                className={classes.authButton}
                size="large"
                color="primary"
                variant="outlined"
                href={`${connection.API_URL}/auth/github`}
              >
                <GithubIcon className={classes.icon} />
                Connect with Github
              </Button>
            </Grid>
            <Grid item xs={12}>
              <Button
                className={classes.authButton}
                size="large"
                color="primary"
                variant="outlined"
                href={`${connection.API_URL}/auth/google`}
              >
                <GoogleIcon className={classes.icon} />
                Connect with Google
              </Button>
            </Grid>
          </Grid>
        </div>
        <VSpace units={4} />
        <Typography className={classes.title} variant="body2" color={"textSecondary"} align="center">
          * We promise to treat your personal details with care
        </Typography>
        <Typography className={classes.title} variant="body2" color={"textSecondary"} align="center">
          ** Manual user authentication coming soon
        </Typography>
      </Container>
    </div>
  </Page>
));
