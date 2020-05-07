import {
  Button,
  Container,
  Grid,
  makeStyles,
  Theme,
  Typography,
} from "@material-ui/core";
import { NextPage } from "next";

import { withApollo } from "../../apollo/withApollo";
import { GithubIcon, GoogleIcon } from "../../components/Icons";
import LinkTypography from "../../components/LinkTypography";
import Page from "../../components/Page";
import VSpace from "../../components/VSpace";
import connection from "../../lib/connection";

const useStyles = makeStyles((theme: Theme) => ({
  authButton: {
    width: "100%",
  },
  authButtons: {
    marginTop: theme.spacing(4),
  },
  icon: {
    fontSize: 24,
    marginRight: theme.spacing(1),
  },
  title: {
    lineHeight: "150%",
    marginBottom: theme.spacing(2),
  },
}));

const AuthPage: NextPage = () => {
  const classes = useStyles();
  return (
    <Page title="Register or Login" contentMarginTop="normal">
      <Container maxWidth="lg">
        <Typography className={classes.title} component="h2" variant="h1" align="center">
          Hello there! Pick an option to sign up or log in
        </Typography>
      </Container>
      <Container maxWidth="sm">
        <div className={classes.authButtons}>
          <Grid container spacing={2} justify="center">
            <Grid item xs={12} md={6}>
              <Button
                className={classes.authButton}
                size="medium"
                color="primary"
                variant="outlined"
                href={`${connection.API_URL}/auth/github`}
              >
                <GithubIcon className={classes.icon} />
                Connect with Github
              </Button>
            </Grid>
            <Grid item xs={12} md={6}>
              <Button
                className={classes.authButton}
                size="medium"
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
          By signing up or logging in you accept our&nbsp;
          <LinkTypography href="https://about.beneath.dev/policies/terms/">Terms of Service</LinkTypography>
          &nbsp;and&nbsp;
          <LinkTypography href="https://about.beneath.dev/policies/privacy/">Privacy Policy</LinkTypography>
        </Typography>
      </Container>
    </Page>
  );
};

export default withApollo(AuthPage);
