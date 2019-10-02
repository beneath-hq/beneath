import clsx from "clsx";
import Link from "next/link";
import React from "react";
import { Query } from "react-apollo";

import {
  Button,
  Container,
  Grid,
  List,
  ListItem,
  ListItemAvatar,
  ListItemSecondaryAction,
  ListItemText,
  makeStyles,
  Paper,
  Typography,
} from "@material-ui/core";

import { EXPLORE_PROJECTS } from "../apollo/queries/project";
import Avatar from "../components/Avatar";
import Loading from "../components/Loading";
import Page from "../components/Page";
import NextMuiLink from "../components/NextMuiLink";
import withMe from "../hocs/withMe";

const useStyles = makeStyles((theme) => ({
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

const Explore = ({ me }) => {
  const loggedIn = !!me;
  const classes = useStyles();
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
                size="large"
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
                size="large"
                color="primary"
                variant="outlined"
                className={clsx(classes.button, classes.primaryButton)}
                href={`/auth`}
                component={NextMuiLink}
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
            ... well, currently the only projects.
          </Typography>
        </Typography>
        <Query query={EXPLORE_PROJECTS}>
          {({ loading, error, data }) => {
            if (loading) {
              return <Loading justify="center" />;
            }
            if (error) {
              return <p>Error: {JSON.stringify(error)}</p>;
            }

            return (
              <Grid container spacing={3} justify="center">
                {data.exploreProjects.map(({ projectID, name, displayName, description, photoURL }) => (
                  <Grid key={projectID} item lg={4} md={6} xs={12}>
                    <Link href={`/project?name=${name}`} as={`/projects/${name}`}>
                      <Paper className={classes.paper}>
                        <Grid container wrap="nowrap" spacing={0}>
                          <Grid item className={classes.avatar}>
                            <Avatar size="list" label={displayName} src={photoURL} />
                          </Grid>
                          <Grid item>
                            <Typography variant="h2">{displayName}</Typography>
                            <Typography color="textSecondary" variant="body2" gutterBottom>
                              /projects/{name}
                            </Typography>
                          </Grid>
                        </Grid>
                        <Typography variant="body1">{description}</Typography>
                      </Paper>
                    </Link>
                  </Grid>
                ))}
              </Grid>
            );
          }}
        </Query>
      </Container>
    </Page>
  );
};

export default withMe(Explore);
