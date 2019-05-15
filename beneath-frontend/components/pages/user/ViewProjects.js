import React from "react";

import Avatar from "@material-ui/core/Avatar";
import List from "@material-ui/core/List";
import ListItem from "@material-ui/core/ListItem";
import ListItemAvatar from "@material-ui/core/ListItemAvatar";
import ListItemText from "@material-ui/core/ListItemText";

import NextMuiLink from "../../NextMuiLink";

const ViewProjects = ({ user }) => {
  return (
    <List>
      {user.projects.map(({ projectId, name, displayName, description, photoUrl }) => (
        <ListItem
          key={projectId}
          component={NextMuiLink} as={`/projects/${name}`} href={`/project?name=${name}`}
          disableGutters button
        >
          {photoUrl && <ListItemAvatar><Avatar alt={displayName} src={photoUrl} /></ListItemAvatar>}
          <ListItemText primary={displayName} secondary={description} />
        </ListItem>
      ))}
    </List>
  );
};

export default ViewProjects;
