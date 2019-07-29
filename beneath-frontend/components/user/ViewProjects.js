import React from "react";

import List from "@material-ui/core/List";
import ListItem from "@material-ui/core/ListItem";
import ListItemAvatar from "@material-ui/core/ListItemAvatar";
import ListItemText from "@material-ui/core/ListItemText";

import Avatar from "../Avatar";
import NextMuiLink from "../NextMuiLink";

const ViewProjects = ({ user }) => {
  return (
    <List>
      {user.projects.map(({ projectID, name, displayName, description, photoURL }) => (
        <ListItem
          key={projectID}
          component={NextMuiLink} as={`/projects/${name}`} href={`/project?name=${name}`}
          disableGutters button
        >
          <ListItemAvatar><Avatar size="list" label={displayName} src={photoURL} /></ListItemAvatar>
          <ListItemText primary={displayName} secondary={description} />
        </ListItem>
      ))}
    </List>
  );
};

export default ViewProjects;
