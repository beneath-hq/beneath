import React from "react";

import List from "@material-ui/core/List";
import ListItem from "@material-ui/core/ListItem";
import ListItemText from "@material-ui/core/ListItemText";
import ListItemAvatar from "@material-ui/core/ListItemAvatar";

import { toURLName } from "../../lib/names";
import Avatar from "../Avatar";
import NextMuiLink from "../NextMuiLink";

const ViewStreams = ({ project }) => {
  return (
    <List>
      {project.streams.map(({ streamID, name, description, external }) => (
        <ListItem
          key={streamID}
          component={NextMuiLink}
          as={`/projects/${toURLName(project.name)}/streams/${toURLName(name)}`}
          href={`/stream?name=${toURLName(name)}&project_name=${toURLName(project.name)}`}
          button
          disableGutters
        >
          <ListItemAvatar>
            <Avatar size="list" label={external ? "Root" : "Derived"} />
          </ListItemAvatar>
          <ListItemText primary={toURLName(name)} secondary={description} />
        </ListItem>
      ))}
    </List>
  );
};

export default ViewStreams;
