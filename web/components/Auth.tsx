import {
  Button,
  Container,
  Grid,
  makeStyles,
  Paper,
  Theme,
  Typography,
  useMediaQuery,
  useTheme,
} from "@material-ui/core";

import { GithubIcon, GoogleIcon } from "components/Icons";
import { Link } from "components/Link";
import { API_URL } from "lib/connection";
import React, { FC } from "react";

const useStyles = makeStyles((theme: Theme) => ({
  title: {
    marginBottom: theme.spacing(4),
    fontSize: "2.5rem",
  },
  list: {
    "&>li": {
      paddingBottom: theme.spacing(0.5),
    },
  },
  loginTitle: {
    marginBottom: theme.spacing(3),
  },
  termsText: {
    paddingTop: theme.spacing(3),
  },
  paper: {
    padding: theme.spacing(4),
    marginBottom: theme.spacing(2),
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

const Auth: FC = () => {
  const classes = useStyles();
  const theme = useTheme();
  const isMd = useMediaQuery(theme.breakpoints.up("md"), { defaultMatches: true });

  return (
    <>
      <Container maxWidth="md">
        <Typography className={classes.title} variant="h1" component="h2">
          Welcome to Beneath
        </Typography>
        <Grid container spacing={3} direction={isMd ? "row" : "column-reverse"}>
          <Grid item xs={12} md={7}>
            <Grid container spacing={3}>
              <Grid item>
                <Typography variant="body1" component="div">
                  <strong>Today, you can use Beneath to:</strong>
                  <ul className={classes.list}>
                    <li>Replay, subscribe and analyze streams</li>
                    <li>Create and share new serverless streams</li>
                    <li>Use the console to browse streams and monitor usage</li>
                    <li>Access streams using Python, SQL, WebSockets and more</li>
                  </ul>
                </Typography>
              </Grid>
              <Grid item>
                <Typography variant="body1" component="div">
                  <strong>To learn more about Beneath:</strong>
                  <ul className={classes.list}>
                    <li>
                      Read about the{" "}
                      <Link href="https://about.beneath.dev/" target="_blank">
                        current&nbsp;features
                      </Link>
                    </li>
                    <li>
                      Learn about the vision and roadmap on{" "}
                      <Link href="https://github.com/beneath-hq/beneath/" target="_blank">
                        GitHub
                      </Link>{" "}
                    </li>
                  </ul>
                </Typography>
              </Grid>
            </Grid>
          </Grid>
          <Grid item xs={12} md={5}>
            <Paper className={classes.paper}>
              <Typography className={classes.loginTitle} align="center" variant="h2" gutterBottom>
                Sign up for free
              </Typography>
              <Grid container spacing={2} justify="center">
                <Grid item xs={12}>
                  <Button
                    className={classes.githubButton}
                    variant="contained"
                    href={`${API_URL}/auth/github`}
                    fullWidth
                  >
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
              <Typography className={classes.termsText} variant="body2" color={"textSecondary"} align="center">
                By signing up or logging in you accept our{" "}
                <Link href="https://about.beneath.dev/policies/terms/" target="_blank">
                  Terms&nbsp;of&nbsp;Service
                </Link>{" "}
                and{" "}
                <Link href="https://about.beneath.dev/policies/privacy/" target="_blank">
                  Privacy&nbsp;Policy
                </Link>
              </Typography>
            </Paper>
          </Grid>
        </Grid>
      </Container>
    </>
  );
};

export default Auth;
