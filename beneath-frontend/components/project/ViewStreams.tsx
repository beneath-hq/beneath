import React, { FC } from "react";

import { List, ListItem, ListItemAvatar, ListItemText, makeStyles, Typography } from "@material-ui/core";

import { ProjectByName_projectByName } from "../../apollo/types/ProjectByName";
import { toURLName } from "../../lib/names";
import Avatar from "../Avatar";
import NextMuiLinkList from "../NextMuiLinkList";

interface ViewStreamsProps {
  project: ProjectByName_projectByName;
}

const useStyles = makeStyles((theme) => ({
  noDataCaption: {
    color: theme.palette.text.secondary,
  },
}));

const ViewStreams: FC<ViewStreamsProps> = ({ project }) => {
  const classes = useStyles();
  return (
    <>
      <List>
        {project.streams.map(({ streamID, name, description, external }) => (
          <ListItem
            key={streamID}
            component={NextMuiLinkList}
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
      {project.streams.length === 0 && (
        <Typography className={classes.noDataCaption} variant="body1" align="center">
          There are no streams in this project
        </Typography>
      )}
    </>
  );
};

export default ViewStreams;
