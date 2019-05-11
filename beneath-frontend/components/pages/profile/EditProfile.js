import React from "react";
import gql from "graphql-tag";
import { Mutation } from "react-apollo";

import Button from "@material-ui/core/Button";
import Moment from "react-moment";
import TextField from "@material-ui/core/TextField";
import Typography from "@material-ui/core/Typography";
import { makeStyles } from "@material-ui/core";

import VSpace from "../../VSpace";

const UPDATE_ME = gql`
  mutation UpdateMe($name: String, $bio: String) {
    updateMe(name: $name, bio: $bio) {
      userId
      name
      bio
      updatedOn
    }
  }
`;

const useStyles = makeStyles((theme) => ({
  submitButton: {
    marginTop: theme.spacing(3),
  },
}));

const EditProfile = ({ me }) => {
  const [values, setValues] = React.useState({
    email: me.email || "",
    name: me.name || "",
    bio: me.bio || "",
    photoUrl: me.photoUrl || "",
  });

  const handleChange = (name) => (event) => {
    setValues({ ...values, [name]: event.target.value });
  };

  const classes = useStyles();
  return (
    <Mutation mutation={UPDATE_ME}>
      {(updateMe, { loading, error }) => (
        <div>
          <form onSubmit={(e) => {
            e.preventDefault();
            updateMe({ variables: { name: values.name, bio: values.bio } });
          }}
          >
            <TextField id="name" label="Name" value={values.name}
              margin="normal" fullWidth required
              onChange={handleChange("name")}
            />
            <TextField id="bio" label="Bio" value={values.bio}
              margin="normal" fullWidth multiline rows={1} rowsMax={3}
              onChange={handleChange("bio")}
            />
            <TextField id="email" label="Email" value={values.email}
              margin="normal" fullWidth disabled
              onChange={handleChange("email")}
            />
            <TextField id="photo-url" label="Photo URL" value={values.photoUrl || ""}
              margin="normal" fullWidth disabled
              onChange={handleChange("photoUrl")}
            />
            <Button type="submit" variant="outlined" color="primary" className={classes.submitButton}
              disabled={
                loading
                || !(values.name && values.name.length >= 4 && values.name.length <= 50)
                || !(values.bio === "" || values.bio.length < 256)
              }>
              Save changes
            </Button>
            {error && (
              <Typography variant="body1" color="error">An error occurred</Typography>
            )}
          </form>
          <VSpace units={2} />
          <Typography variant="subtitle1" color="textSecondary" gutterBottom>
            You signed up <Moment fromNow date={me.createdOn} /> and
            last updated your profile <Moment fromNow date={me.updatedOn} />.
          </Typography>
        </div>
      )}
    </Mutation>
  );
};

export default EditProfile;
