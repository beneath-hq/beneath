import React from "react";
import isEmail from "validator/lib/isEmail";
import { Mutation } from "react-apollo";

import Avatar from "@material-ui/core/Avatar";
import Button from "@material-ui/core/Button";
import DeleteIcon from "@material-ui/icons/Delete";
import Grid from "@material-ui/core/Grid";
import IconButton from "@material-ui/core/IconButton";
import List from "@material-ui/core/List";
import ListItem from "@material-ui/core/ListItem";
import ListItemAvatar from "@material-ui/core/ListItemAvatar";
import ListItemText from "@material-ui/core/ListItemText";
import ListItemSecondaryAction from "@material-ui/core/ListItemSecondaryAction";
import TextField from "@material-ui/core/TextField";
import Typography from "@material-ui/core/Typography";
import { makeStyles } from "@material-ui/core";

import Loading from "../../Loading";
import NextMuiLink from "../../NextMuiLink";
import VSpace from "../../VSpace";

import { QUERY_PROJECT, ADD_MEMBER, REMOVE_MEMBER } from "../../../queries/project";

const useStyles = makeStyles((theme) => ({
}));

const ManageMembers = ({ project, editable }) => {
  return (
    <React.Fragment>
      <ViewMembers project={project} editable={editable} />
      {editable && <VSpace units={2} />}
      {editable && <AddMember project={project} />}
    </React.Fragment>
  );
};

export default ManageMembers;

const ViewMembers = ({ project, editable }) => {
  const classes = useStyles();
  return (
    <List>
      {project.users.map(({ userId, username, name, photoUrl }) => (
        <ListItem
          key={userId}
          component={NextMuiLink} as={`/users/${userId}`} href={`/user?id=${userId}`}
          disableGutters button
        >
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
                    data: { project: { ...project, users: project.users.filter((user) => user.userId !== userId) } },
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
};

const AddMember = ({ project }) => {
  const [email, setEmail] = React.useState("");
  const classes = useStyles();
  return (
    <Mutation mutation={ADD_MEMBER} update={(cache, { data: { addUserToProject } }) => {
      const user = addUserToProject;
      const query = cache.readQuery({ query: QUERY_PROJECT, variables: { name: project.name } });
      cache.writeQuery({
        query: QUERY_PROJECT,
        variables: { name: project.name },
        data: { project: { ...query.project, users: query.project.users.concat([user]) } },
      });
    }} onCompleted={() => setEmail("")}>
      {(addUserToProject, { loading, error }) => (
        <form onSubmit={(e) => {
          e.preventDefault();
          addUserToProject({ variables: { email, projectId: project.projectId } });
        }}>
          <Grid container alignItems={"center"} spacing={2}>
            <Grid item sm={true}>
              <TextField id="email" type="email" label="Email" value={email}
                fullWidth disabled={loading} error={!!error} helperText={error && error.graphQLErrors[0].message}
                onChange={(event) => setEmail(event.target.value)}
              />
            </Grid>
            <Grid item>
              <Button type="submit" variant="outlined" color="primary" disabled={loading || !isEmail(email)}>
                Add member
              </Button>
            </Grid>
          </Grid>
        </form>
      )}
    </Mutation>
  );
};
