import React from "react";
import { Mutation } from "react-apollo";

import Button from "@material-ui/core/Button";
import Moment from "react-moment";
import TextField from "@material-ui/core/TextField";
import Typography from "@material-ui/core/Typography";
import { makeStyles } from "@material-ui/core";
import isUrl from "validator/lib/isUrl";

import VSpace from "../../VSpace";

import { UPDATE_PROJECT } from "../../../queries/project";

const useStyles = makeStyles((theme) => ({
  submitButton: {
    marginTop: theme.spacing(3),
  },
}));

const EditProject = ({ project }) => {
  const [values, setValues] = React.useState({
    displayName: project.displayName || "",
    site: project.site || "",
    description: project.description || "",
    photoUrl: project.photoUrl || "",
  });

  const handleChange = (name) => (event) => {
    setValues({ ...values, [name]: event.target.value });
  };

  const classes = useStyles();
  return (
    <Mutation mutation={UPDATE_PROJECT}>
      {(updateProject, { loading, error }) => (
        <div>
          <form onSubmit={(e) => {
            e.preventDefault();
            updateProject({ variables: { projectId: project.projectId, ...values } });
          }}
          >
            <TextField id="name" label="Name" value={project.name}
              margin="normal" fullWidth disabled
            />
            <TextField id="displayName" label="Display Name" value={values.displayName}
              margin="normal" fullWidth required
              onChange={handleChange("displayName")}
            />
            <TextField id="site" label="Site" value={values.site}
              margin="normal" fullWidth
              onChange={handleChange("site")}
            />
            <TextField id="description" label="Description" value={values.description}
              margin="normal" fullWidth
              onChange={handleChange("description")}
            />
            <TextField id="photoUrl" label="Photo Url" value={values.photoUrl}
              margin="normal" fullWidth
              onChange={handleChange("photoUrl")}
            />
            <Button type="submit" variant="outlined" color="primary" className={classes.submitButton}
              disabled={
                loading
                || !(values.displayName && values.displayName.length >= 4 && values.displayName.length <= 40)
                || !(values.site === "" || isUrl(values.site))
                || !(values.description === "" || values.description.length < 256)
                || !(values.photoUrl === "" || isUrl(values.photoUrl))
              }>
              Save changes
            </Button>
            {error && (
              <Typography variant="body1" color="error">An error occurred</Typography>
            )}
          </form>
          <VSpace units={2} />
          <Typography variant="subtitle1" color="textSecondary">
            The project was created <Moment fromNow date={project.createdOn} /> and
            last updated <Moment fromNow date={project.updatedOn} />.
          </Typography>
        </div>
      )}
    </Mutation>
  );
};

export default EditProject;
