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

const Auth: FC = () => {
  const classes = useStyles();
  const theme = useTheme();
  const isMd = useMediaQuery(theme.breakpoints.up("md"));

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
                  Today, you can use Beneath to:
                  <ul>
                    <li>Replay, subscribe and query streams</li>
                    <li>Create and share new streams</li>
                    <li>Explore streams and browse usage</li>
                  </ul>
                </Typography>
              </Grid>
              <Grid item>
                <Typography variant="body1" component="div">
                  To learn more about the beta:
                  <ul>
                    <li>
                      Read about the <Link href="https://about.beneath.dev/beta/features/">current&nbsp;features</Link>
                    </li>
                  </ul>
                </Typography>
              </Grid>
              <Grid item>
                <Typography variant="body1" component="div">
                  Beneath is just getting started. To learn about what's next:
                  <ul>
                    <li>
                      Check out the <Link href="https://github.com/beneath-hq/beneath/">GitHub</Link>
                    </li>
                  </ul>
                </Typography>
              </Grid>
            </Grid>
          </Grid>
          <Grid item xs={12} md={5}>
            <Paper className={classes.paper}>
              <Typography align="center" gutterBottom>
                Pick an option to sign up or log in
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
            </Paper>
            <Typography className={classes.termsText} variant="body2" color={"textSecondary"} align="center">
              By signing up or logging in you accept our{" "}
              <Link href="https://about.beneath.dev/policies/terms/">Terms&nbsp;of&nbsp;Service</Link> and{" "}
              <Link href="https://about.beneath.dev/policies/privacy/">Privacy&nbsp;Policy</Link>
            </Typography>
          </Grid>
        </Grid>
      </Container>
    </>
  );
};

export default Auth;
