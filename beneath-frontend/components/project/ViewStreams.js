import React from "react";

import List from "@material-ui/core/List";
import ListItem from "@material-ui/core/ListItem";
import ListItemText from "@material-ui/core/ListItemText";
import ListItemAvatar from "@material-ui/core/ListItemAvatar";

import Avatar from "../Avatar";
import NextMuiLink from "../NextMuiLink";

const ViewStreams = ({ project }) => {
  return (
    <List>
      {project.streams.map(({ streamID, name, description, external }) => (
        <ListItem
          key={streamID}
          component={NextMuiLink} as={`/projects/${project.name}/streams/${name}`} href={`/stream?name=${name}&project_name=${project.name}`}
          button
          disableGutters
        >
          <ListItemAvatar><Avatar size="list" label={external ? "External" : "Internal"} /></ListItemAvatar>
          <ListItemText primary={name} secondary={description} />
        </ListItem>
      ))}
    </List>
  );
};

export default ViewStreams;
