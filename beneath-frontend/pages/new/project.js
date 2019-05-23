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
import { QUERY_ME } from "../../queries/user";

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
    <Page title="New Project" sidebar={<ExploreSidebar />} maxWidth="md" contentMarginTop="normal">
      <Mutation mutation={NEW_PROJECT} 
        refetchQueries={[{ query: QUERY_ME }]}
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

          const isNameError = error && error.message.match(/IDX_UQ_PROJECTS_NAME/);

          return (
            <form onSubmit={onSubmit}>
              <Typography component="h2" variant="h2" gutterBottom>Create project</Typography>
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
                margin="normal" fullWidth multiline required
                onChange={handleChange("description")}
              />
              <Button type="submit" variant="outlined" color="primary" className={classes.submitButton}
                disabled={
                  loading
                  || !(values.name.match(/^[_a-z][_\-a-z0-9]*$/))
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
