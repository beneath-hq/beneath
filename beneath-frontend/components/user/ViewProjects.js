import React from "react";

import List from "@material-ui/core/List";
import ListItem from "@material-ui/core/ListItem";
import ListItemAvatar from "@material-ui/core/ListItemAvatar";
import ListItemText from "@material-ui/core/ListItemText";

import { toURLName } from "../../lib/names";
import Avatar from "../Avatar";
import NextMuiLink from "../NextMuiLink";

const ViewProjects = ({ user }) => {
  return (
    <List>
      {user.projects.map(({ projectID, name, displayName, description, photoURL }) => (
        <ListItem
          key={projectID}
          component={NextMuiLink} as={`/projects/${toURLName(name)}`} href={`/project?name=${toURLName(name)}`}
          disableGutters button
        >
          <ListItemAvatar><Avatar size="list" label={displayName || name} src={photoURL} /></ListItemAvatar>
          <ListItemText primary={displayName || name} secondary={description} />
        </ListItem>
      ))}
    </List>
  );
};

export default ViewProjects;
