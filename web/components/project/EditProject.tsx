import { useMutation } from "@apollo/react-hooks";
import React, { FC } from "react";
import Moment from "react-moment";
import validator from "validator";

import { Button, makeStyles, TextField, Typography } from "@material-ui/core";

import { STAGE_PROJECT } from "../../apollo/queries/project";
import { ProjectByOrganizationAndName_projectByOrganizationAndName } from "../../apollo/types/ProjectByOrganizationAndName";
import { UpdateProject, UpdateProjectVariables } from "../../apollo/types/UpdateProject";
import VSpace from "../VSpace";

const useStyles = makeStyles((theme) => ({
  submitButton: {
    marginTop: theme.spacing(3),
  },
}));

interface EditProjectProps {
  project: ProjectByOrganizationAndName_projectByOrganizationAndName;
}

const EditProject: FC<EditProjectProps> = ({ project }) => {
  const [values, setValues] = React.useState({
    displayName: project.displayName || "",
    site: project.site || "",
    description: project.description || "",
    photoURL: project.photoURL || "",
  });

  const [updateProject, { loading, error }] = useMutation<UpdateProject, UpdateProjectVariables>(STAGE_PROJECT);

  const handleChange = (name: string) => (event: any) => {
    setValues({ ...values, [name]: event.target.value });
  };

  const classes = useStyles();
  return (
    <div>
      <form
        onSubmit={(e) => {
          e.preventDefault();
          updateProject({ variables: { projectID: project.projectID, ...values } });
        }}
      >
        <TextField id="name" label="Name" value={project.name} margin="normal" fullWidth disabled />
        <TextField
          id="displayName"
          label="Display Name"
          value={values.displayName}
          margin="normal"
          fullWidth
          onChange={handleChange("displayName")}
        />
        <TextField
          id="site"
          label="Site"
          value={values.site}
          margin="normal"
          fullWidth
          onChange={handleChange("site")}
        />
        <TextField
          id="description"
          label="Description"
          value={values.description}
          margin="normal"
          fullWidth
          onChange={handleChange("description")}
        />
        <TextField
          id="photoURL"
          label="Photo Url"
          value={values.photoURL}
          margin="normal"
          fullWidth
          onChange={handleChange("photoURL")}
        />
        <Button
          type="submit"
          variant="outlined"
          color="primary"
          className={classes.submitButton}
          disabled={
            loading ||
            !(values.displayName.length <= 40) ||
            !(values.site === "" || validator.isURL(values.site)) ||
            !(values.description === "" || values.description.length < 256) ||
            !(values.photoURL === "" || validator.isURL(values.photoURL))
          }
        >
          Save changes
        </Button>
        {error && (
          <Typography variant="body1" color="error">
            An error occurred: {JSON.stringify(error)}
          </Typography>
        )}
      </form>
      <VSpace units={2} />
      <Typography variant="subtitle1" color="textSecondary">
        The project was created <Moment fromNow date={project.createdOn} /> and last updated{" "}
        <Moment fromNow date={project.updatedOn} />.
      </Typography>
    </div>
  );
};

export default EditProject;
