import React, { Component } from "react";
import Router from "next/router";
import { Mutation } from "react-apollo";

import Button from "@material-ui/core/Button";
import TextField from "@material-ui/core/TextField";
import Typography from "@material-ui/core/Typography";
import { makeStyles } from "@material-ui/core";

import ExploreSidebar from "../../components/ExploreSidebar";
import Page from "../../components/Page";

import { NEW_PROJECT } from "../../queries/project";
import { QUERY_USER } from "../../queries/user";

const useStyles = makeStyles((theme) => ({
  submitButton: {
    marginTop: theme.spacing(3),
  },
}));

const NewProjectPage = () => {
  const [values, setValues] = React.useState({
    name: "",
    displayName: "",
    description: "",
  });

  const handleChange = (name) => (event) => {
    setValues({ ...values, [name]: event.target.value });
  };

  const classes = useStyles();
  return (
    <Page title="New Project" sidebar={<ExploreSidebar />}>
      <Mutation mutation={NEW_PROJECT} 
        update={(cache, { data: { createProject } }) => {
          // TODO: Update QUERY_USER (not very important). The below should work, but fails
          // const project = createProject;
          // const userId = project.users[0].userId;
          // console.log(cache);
          // const query = cache.readQuery({ query: QUERY_USER, variables: { userId } });
          // if (query) {
          //   cache.writeQuery({
          //     query: QUERY_USER,
          //     variables: { userId },
          //     data: {
          //       user: {
          //         ...query.user,
          //         projects: query.user.projects.concat([{
          //           projectId: project.projectId,
          //           name: project.name,
          //           displayName: project.displayName,
          //           description: project.description,
          //           photoUrl: project.photoUrl,
          //         }])
          //       }
          //     },
          //   });
          // }
        }}
        onCompleted={({ createProject }) => {
          const project = createProject;
          Router.push(`/project?name=${project.name}`, `/projects/${project.name}`);
        }}
      >
        {(newProject, { loading, error }) => {
          const onSubmit = (e) => {
            e.preventDefault();
            newProject({ variables: values });
          };

          const isNameError = error && error.message.match(/duplicate key/);

          return (
            <form onSubmit={onSubmit}>
              <TextField id="name" label="Name" value={values.name}
                margin="normal" fullWidth required
                error={isNameError} helperText={isNameError && "Project name already taken"}
                onChange={handleChange("name")}
              />
              <TextField id="displayName" label="Display Name" value={values.displayName}
                margin="normal" fullWidth required
                onChange={handleChange("displayName")}
              />
              <TextField id="description" label="Description" value={values.description}
                margin="normal" fullWidth required
                onChange={handleChange("description")}
              />
              <Button type="submit" variant="outlined" color="primary" className={classes.submitButton}
                disabled={
                  loading
                  || !(values.name.match(/[_a-z][_\-a-z0-9]*/)) 
                  || !(values.name && values.name.length >= 3 && values.name.length <= 16)
                  || !(values.displayName && values.displayName.length >= 3 && values.displayName.length <= 40)
                  || !(values.description !== "" && values.description.length <= 255)
                }>
                Create project
              </Button>
              {error && !isNameError && (
                <Typography variant="body1" color="error">An error occurred</Typography>
              )}
            </form>
          );
        }}
      </Mutation>
    </Page>
  );
};

export default NewProjectPage;
