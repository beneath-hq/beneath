import React from "react";

import Avatar from "@material-ui/core/Avatar";
import List from "@material-ui/core/List";
import ListItem from "@material-ui/core/ListItem";
import ListItemText from "@material-ui/core/ListItemText";
import ListItemAvatar from "@material-ui/core/ListItemAvatar";

import NextMuiLink from "../../NextMuiLink";

const ViewStreams = ({ project }) => {
  return (
    <List>
      {project.streams.map(({ streamId, name, description, external }) => (
        <ListItem
          key={streamId}
          component={NextMuiLink} as={`/projects/${project.name}/streams/${name}`} href={`/stream?name=${name}&project_name=${project.name}`}
          button
          disableGutters
        >
          <ListItemAvatar><Avatar>{external ? "E" : "I"}</Avatar></ListItemAvatar>
          <ListItemText primary={name} secondary={description} />
        </ListItem>
      ))}
    </List>
  );
};

export default ViewStreams;
