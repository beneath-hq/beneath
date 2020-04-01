import React, { FC } from "react";
import { useQuery } from "@apollo/react-hooks";
import {
  Container,
  Grid,
  makeStyles,
  Paper,
  Theme,
  Typography,
} from "@material-ui/core";
import Link from "next/link";
import { toURLName } from "../../lib/names";
import Avatar from "../Avatar";
import { EXPLORE_PROJECTS } from "../../apollo/queries/project";
import { ExploreProjects } from "../../apollo/types/ExploreProjects";
import ErrorPage from "../ErrorPage";
import Loading from "../Loading";

const useStyles = makeStyles((theme: Theme) => ({
  sectionHeader: {
    fontSize: theme.typography.pxToRem(24),
    marginBottom: theme.spacing(6),
  },
  sectionSubHeader: {
    marginTop: theme.spacing(2),
  },
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

const TopProjects: FC = () => {
  const classes = useStyles();

  const { loading, error, data } = useQuery<ExploreProjects>(EXPLORE_PROJECTS);
  if (loading) {
    return <Loading justify="center" />;
  }

  if (error || !data) {
    return <ErrorPage apolloError={error} />;
  }

  return (
    <React.Fragment>
      <Container maxWidth="lg">
        <Typography className={classes.sectionHeader} variant="h3" gutterBottom align="center">
          Top projects
                <Typography className={classes.sectionSubHeader} variant="body2" gutterBottom align="center">
            Discover the top public data streams on Beneath.
                </Typography>
        </Typography>
        <Grid container spacing={3} justify="center">
          {data.exploreProjects.map(({ projectID, name, displayName, description, photoURL, organization }) => (
            <Grid key={projectID} item lg={4} md={6} xs={12}>
              <Link href={`/${toURLName(organization.name)}/${toURLName(name)}`}>
                <Paper className={classes.paper}>
                  <Grid container wrap="nowrap" spacing={0}>
                    <Grid item className={classes.avatar}>
                      <Avatar size="list" label={displayName || name} src={photoURL || undefined} />
                    </Grid>
                    <Grid item>
                      <Typography variant="h2">{displayName || toURLName(name)}</Typography>
                      <Typography color="textSecondary" variant="body2" gutterBottom>
                        /{toURLName(organization.name)}/{toURLName(name)}
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
    </React.Fragment>
  )
}

export default TopProjects