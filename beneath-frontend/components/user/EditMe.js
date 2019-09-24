import React from "react";
import { Query, Mutation } from "react-apollo";

import Button from "@material-ui/core/Button";
import Moment from "react-moment";
import TextField from "@material-ui/core/TextField";
import Typography from "@material-ui/core/Typography";
import { makeStyles } from "@material-ui/core";

import Loading from "../Loading";
import VSpace from "../VSpace";

import { QUERY_ME, UPDATE_ME } from "../../apollo/queries/user";

const useStyles = makeStyles((theme) => ({
  submitButton: {
    marginTop: theme.spacing(3),
  },
}));

const EditMe = () => {
  return (
    <Query query={QUERY_ME}>
      {({ loading, error, data }) => {
        if (loading) {
          return <Loading justify="center" />;
        }
        if (error) {
          return <p>Error: {JSON.stringify(error)}</p>;
        }

        let { me } = data;
        return <EditMeForm me={me} />;
      }}
    </Query>
  );
};

export default EditMe;

const EditMeForm = ({ me }) => {
  const [values, setValues] = React.useState({
    username: me.user.username || "",
    email: me.email || "",
    name: me.user.name || "",
    bio: me.user.bio || "",
    photoURL: me.user.photoURL || "",
  });

  const handleChange = (name) => (event) => {
    setValues({ ...values, [name]: event.target.value });
  };

  const classes = useStyles();
  return (
    <Mutation mutation={UPDATE_ME}>
      {(updateMe, { loading, error }) => (
        <div>
          <form
            onSubmit={(e) => {
              e.preventDefault();
              updateMe({ variables: { username: values.username, name: values.name, bio: values.bio } });
            }}
          >
            <TextField
              id="username"
              label="Username"
              value={values.username}
              margin="normal"
              fullWidth
              required
              onChange={handleChange("username")}
            />
            <TextField
              id="name"
              label="Name"
              value={values.name}
              margin="normal"
              fullWidth
              required
              onChange={handleChange("name")}
            />
            <TextField
              id="bio"
              label="Bio"
              value={values.bio}
              margin="normal"
              fullWidth
              multiline
              rows={1}
              rowsMax={3}
              onChange={handleChange("bio")}
            />
            <TextField
              id="email"
              label="Email"
              value={values.email}
              margin="normal"
              fullWidth
              disabled
              onChange={handleChange("email")}
            />
            <TextField
              id="photo-url"
              label="Photo URL"
              value={values.photoURL || ""}
              margin="normal"
              fullWidth
              disabled
              onChange={handleChange("photoURL")}
            />
            <Button
              type="submit"
              variant="outlined"
              color="primary"
              className={classes.submitButton}
              disabled={
                loading ||
                !(
                  values.username &&
                  values.username.length >= 3 &&
                  values.username.length <= 40 &&
                  values.username.match(/^[_a-z][_a-z0-9]+$/)
                ) ||
                !(values.name && values.name.length >= 4 && values.name.length <= 50) ||
                !(values.bio === "" || values.bio.length < 256)
              }
            >
              Save changes
            </Button>
            {error && (
              <Typography variant="body1" color="error">
                {isUsernameError(error) ? "Username already taken" : `An error occurred: ${JSON.stringify(error)}`}
              </Typography>
            )}
          </form>
          <VSpace units={2} />
          <Typography variant="subtitle1" color="textSecondary">
            You signed up <Moment fromNow date={me.user.createdOn} /> and last updated your profile{" "}
            <Moment fromNow date={me.updatedOn} />.
          </Typography>
        </div>
      )}
    </Mutation>
  );
};

const isUsernameError = (error) => {
  return error.message.match(/duplicate key value violates unique constraint/);
};
