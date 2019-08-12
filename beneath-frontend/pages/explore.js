import React from "react";
import { Query } from "react-apollo";

import List from "@material-ui/core/List";
import ListItem from "@material-ui/core/ListItem";
import ListItemAvatar from "@material-ui/core/ListItemAvatar";
import ListItemText from "@material-ui/core/ListItemText";
import ListItemSecondaryAction from "@material-ui/core/ListItemSecondaryAction";
import { makeStyles } from "@material-ui/core/styles";
import Typography from "@material-ui/core/Typography";

import { EXPLORE_PROJECTS } from "../apollo/queries/project";
import Avatar from "../components/Avatar";
import Loading from "../components/Loading";
import NextMuiLink from "../components/NextMuiLink";
import Page from "../components/Page";

const useStyles = makeStyles((theme) => ({
  title: {
    marginTop: theme.spacing(4),
    marginBottom: theme.spacing(4),
  },
}));

export default () => {
  const classes = useStyles();
  return (
    <Page title="Explore">
      <div className="section">
        <Typography className={classes.title} variant="h1" gutterBottom>
          Explore Projects
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
              <List>
                {data.exploreProjects.map(({ projectID, name, displayName, description, photoURL }) => (
                  <ListItem
                    key={projectID}
                    component={NextMuiLink}
                    as={`/projects/${name}`}
                    href={`/project?name=${name}`}
                    disableGutters
                    button
                  >
                    <ListItemAvatar>
                      <Avatar size="list" label={displayName} src={photoURL} />
                    </ListItemAvatar>
                    <ListItemText primary={displayName} secondary={description} />
                    <ListItemSecondaryAction>
                      <Typography color="textSecondary" variant="body2">
                        /projects/{name}
                      </Typography>
                    </ListItemSecondaryAction>
                  </ListItem>
                ))}
              </List>
            );
          }}
        </Query>
      </div>
    </Page>
  );
}
