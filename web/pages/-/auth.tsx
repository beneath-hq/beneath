import {
  Button,
  Container,
  Grid,
  makeStyles,
  Paper,
  Theme,
  Typography,
} from "@material-ui/core";
import { NextPage } from "next";

import { withApollo } from "apollo/withApollo";
import { GithubIcon, GoogleIcon } from "components/Icons";
import { Link } from "components/Link";
import Page from "components/Page";
import { API_URL } from "lib/connection";

const useStyles = makeStyles((theme: Theme) => ({
  title: {
    marginBottom: theme.spacing(4),
    fontSize: "2.5rem",
  },
  termsText: {
    marginTop: theme.spacing(2),
    padding: theme.spacing(2),
  },
  paper: {
    padding: theme.spacing(4),
  },
  icon: {
    fontSize: 18,
    marginRight: theme.spacing(1),
  },
  githubButton: {
    color: "black",
    backgroundColor: "white",
  },
  googleButton: {
    color: "white",
    backgroundColor: "#4285F4",
    "&:hover": {
      backgroundColor: "#3070E0",
    },
  },
}));

const AuthPage: NextPage = () => {
  const classes = useStyles();
  return (
    <Page title="Welcome to Beneath" contentMarginTop="normal">
      <Typography className={classes.title} variant="h1" component="h2" align="center">
        Welcome to Beneath
      </Typography>
      <Container maxWidth="xs">
        <Paper className={classes.paper}>
          <Typography align="center" gutterBottom>
            Pick an option to sign up or log in
          </Typography>
          <Grid container spacing={2} justify="center">
            <Grid item xs={12}>
              <Button className={classes.githubButton} variant="contained" href={`${API_URL}/auth/github`} fullWidth>
                <GithubIcon className={classes.icon} />
                Connect with Github
              </Button>
            </Grid>
            <Grid item xs={12}>
              <Button
                className={classes.googleButton}
                color="primary"
                variant="contained"
                href={`${API_URL}/auth/google`}
                fullWidth
              >
                <GoogleIcon className={classes.icon} />
                Connect with Google
              </Button>
            </Grid>
          </Grid>
        </Paper>
        <Typography className={classes.termsText} variant="body2" color={"textSecondary"} align="center">
          By signing up or logging in you accept our{" "}
          <Link href="https://about.beneath.dev/policies/terms/">Terms&nbsp;of&nbsp;Service</Link>
          {" "}and{" "}
          <Link href="https://about.beneath.dev/policies/privacy/">Privacy&nbsp;Policy</Link>
        </Typography>
      </Container>
    </Page>
  );
};

export default withApollo(AuthPage);
