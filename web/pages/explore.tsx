import { useQuery } from "@apollo/react-hooks";
import clsx from "clsx";
import { NextPage } from "next";
import Link from "next/link";
import React from "react";

import {
  Button,
  Container,
  Grid,
  makeStyles,
  Paper,
  Theme,
  Typography,
} from "@material-ui/core";

import { EXPLORE_PROJECTS } from "../apollo/queries/project";
import { ExploreProjects } from "../apollo/types/ExploreProjects";
import { withApollo } from "../apollo/withApollo";
import Avatar from "../components/Avatar";
import ErrorPage from "../components/ErrorPage";
import Loading from "../components/Loading";
import NextMuiLinkList from "../components/NextMuiLinkList";
import Page from "../components/Page";
import useMe from "../hooks/useMe";
import { toURLName } from "../lib/names";

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
  exploreTitle: {
    fontSize: theme.typography.pxToRem(24),
    marginBottom: theme.spacing(8),
  },
  exploreTitleJoke: {
    marginTop: theme.spacing(2),
  },
  button: {},
  primaryButton: {},
  secondaryButton: {},
  avatar: {
    marginRight: theme.spacing(1.5),
    marginBottom: theme.spacing(1.5),
  },
  paper: {
    cursor: "pointer",
    padding: theme.spacing(2.5),
    height: "100%",
    "&:hover": {
      boxShadow: theme.shadows[10],
    },
  },
}));

const Explore: NextPage = () => {
  const classes = useStyles();
  const me = useMe();
  const loggedIn = !!me;

  const {loading, error, data} = useQuery<ExploreProjects>(EXPLORE_PROJECTS);
  if (loading) {
    return <Loading justify="center" />;
  }

  if (error || !data) {
    return <ErrorPage apolloError={error} />;
  }

  return (
    <Page title="Explore" contentMarginTop="hero">
      <Container maxWidth="lg">
        <Typography className={classes.title} component="h1" variant="h1" align="center" gutterBottom>
          Become a Blockchain Analytics Power User
        </Typography>
      </Container>
      <Container maxWidth="md">
        <Typography className={classes.subtitle} component="h2" variant="subtitle1" align="center" gutterBottom>
          Beneath is a full blockchain data science platform. From real-time data extraction to analytics development,
          deployment, integration and public sharing.
        </Typography>
        {loggedIn && (
          <Grid className={classes.buttons} container spacing={2} justify="center">
            <Grid item>
              <Button
                size="medium"
                color="default"
                variant="outlined"
                className={clsx(classes.button, classes.secondaryButton)}
                href={`https://about.beneath.network/docs/`}
              >
                Docs
              </Button>
            </Grid>
            <Grid item>
              <Button
                size="medium"
                color="primary"
                variant="outlined"
                className={clsx(classes.button, classes.primaryButton)}
                href={`mailto:contact@beneath.network`}
              >
                Get in touch
              </Button>
            </Grid>
          </Grid>
        )}
        {!loggedIn && (
          <Grid className={classes.buttons} container spacing={2} justify="center">
            <Grid item>
              <Button
                size="medium"
                color="default"
                variant="outlined"
                className={clsx(classes.button, classes.secondaryButton)}
                href={`https://about.beneath.network/docs/`}
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
                href={`/auth`}
                component={NextMuiLinkList}
              >
                Create Account
              </Button>
            </Grid>
          </Grid>
        )}
      </Container>
      <Container maxWidth="lg">
        <Typography className={classes.exploreTitle} variant="h3" gutterBottom align="center">
          Top projects
          <Typography className={classes.exploreTitleJoke} variant="body2" gutterBottom align="center">
            ... well, currently the only projects
          </Typography>
        </Typography>
        <Grid container spacing={3} justify="center">
          {data.exploreProjects.map(({ projectID, name, displayName, description, photoURL }) => (
            <Grid key={projectID} item lg={4} md={6} xs={12}>
              <Link href={`/project?name=${toURLName(name)}`} as={`/projects/${toURLName(name)}`}>
                <Paper className={classes.paper}>
                  <Grid container wrap="nowrap" spacing={0}>
                    <Grid item className={classes.avatar}>
                      <Avatar size="list" label={displayName || name} src={photoURL || undefined} />
                    </Grid>
                    <Grid item>
                      <Typography variant="h2">{displayName || toURLName(name)}</Typography>
                      <Typography color="textSecondary" variant="body2" gutterBottom>
                        /projects/{toURLName(name)}
                      </Typography>
                    </Grid>
                  </Grid>
                  <Typography variant="body1">{description}</Typography>
                </Paper>
              </Link>
            </Grid>
          ))}
        </Grid>
      </Container>
    </Page>
  );
};

export default withApollo(Explore);
