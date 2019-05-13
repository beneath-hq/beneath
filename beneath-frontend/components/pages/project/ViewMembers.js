import React from "react";
import { Mutation } from "react-apollo";

import Avatar from "@material-ui/core/Avatar";
import DeleteIcon from "@material-ui/icons/Delete";
import IconButton from "@material-ui/core/IconButton";
import List from "@material-ui/core/List";
import ListItem from "@material-ui/core/ListItem";
import ListItemAvatar from "@material-ui/core/ListItemAvatar";
import ListItemText from "@material-ui/core/ListItemText";
import ListItemSecondaryAction from "@material-ui/core/ListItemSecondaryAction";

import Loading from "../../Loading";

import { QUERY_PROJECT, REMOVE_MEMBER } from "../../../queries/project";

const ViewMembers = ({ project, editable }) => (
  <List>
    {project.users.map(({ userId, username, name, photoUrl }) => (
      <ListItem key={userId} disableGutters>
        <ListItemAvatar><Avatar alt={name} src={photoUrl} /></ListItemAvatar>
        <ListItemText primary={name} secondary={username} />
        {editable && (project.users.length > 1) && (
          <ListItemSecondaryAction>
            <Mutation mutation={REMOVE_MEMBER} update={(cache, { data: { removeUserFromProject } }) => {
              const projectName = project.name;
              if (removeUserFromProject) {
                const { project } = cache.readQuery({ query: QUERY_PROJECT, variables: { name: projectName } });
                cache.writeQuery({
                  query: QUERY_PROJECT,
                  variables: { name: projectName },
                  data: { project: { ...project, users: project.users.filter((user) => user.userId !== userId) }},
                });
              }
            }}>
              {(removeUserFromProject, { loading, error }) => (
                <IconButton edge="end" aria-label="Delete" onClick={() => {
                  removeUserFromProject({ variables: { projectId: project.projectId, userId: userId } });
                }}>
                  {loading ? <Loading size={20} /> : <DeleteIcon />}
                </IconButton>
              )}
            </Mutation>
          </ListItemSecondaryAction>
        )}
      </ListItem>
    ))}
  </List>
);

export default ViewMembers;
