import clsx from "clsx";
import React, { FC } from "react";

import { Button, Container, Grid, makeStyles, Theme, Typography } from "@material-ui/core";

import TopProjects from "../../components/terminal/TopProjects";
import NextMuiLinkList from "../NextMuiLinkList";

const useStyles = makeStyles((theme: Theme) => ({
  title: {
    fontSize: theme.typography.pxToRem(52),
    marginBottom: theme.spacing(5),
  },
  subtitle: {
    fontSize: theme.typography.pxToRem(20),
    marginBottom: theme.spacing(6),
  },
  buttons: {
    marginBottom: theme.spacing(6),
  },
  button: {},
  primaryButton: {},
  secondaryButton: {},
}));

const Welcome: FC = () => {
  const classes = useStyles();

  return (
    <>
      <Container maxWidth="lg">
        <Typography className={classes.title} component="h1" variant="h1" align="center" gutterBottom>
          Welcome to the Beneath Data Terminal
        </Typography>
      </Container>
      <Container maxWidth="md">
        <Typography className={classes.subtitle} component="h2" variant="subtitle1" align="center" gutterBottom>
          Discover real-time data streams that have been shared with the world.
          Consume public data or create your own stream in minutes.
        </Typography>
        <Grid className={classes.buttons} container spacing={2} justify="center">
          <Grid item>
            <Button
              size="medium"
              color="default"
              variant="outlined"
              className={clsx(classes.button, classes.secondaryButton)}
              href={`https://about.beneath.dev/docs/`}
            >
              Overview
            </Button>
          </Grid>
          <Grid item>
            <Button
              size="medium"
              color="primary"
              variant="outlined"
              className={clsx(classes.button, classes.primaryButton)}
              href={`/-/auth`}
              component={NextMuiLinkList}
            >
              Create Account
            </Button>
          </Grid>
        </Grid>
      </Container>
      <TopProjects/>
    </>
  )
}

export default Welcome;
