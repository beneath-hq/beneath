import { withRouter } from "next/router";
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

const EditMe = ({ router }) => {
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
        return <EditMeForm router={router} me={me} />;
      }}
    </Query>
  );
};

export default withRouter(EditMe);

const EditMeForm = ({ me, router }) => {
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
    <Mutation
      mutation={UPDATE_ME}
      onCompleted={(data) => {
        if (data.updateMe) {
          const username = data.updateMe.user.username;
          if (username !== router.query.name) {
            const href = `/user?name=${username}&tab=edit`;
            const as = `/users/${username}/edit`;
            router.replace(href, as, { shallow: true });
          }
        }
      }}
    >
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
              helperText={
                !validateUsername(values.username)
                  ? "Lowercase letters, numbers and underscores allowed (minimum 3 characters)"
                  : undefined
              }
              error={!validateUsername(values.username)}
              fullWidth
              required
              onChange={handleChange("username")}
            />
            <TextField
              id="name"
              label="Name"
              value={values.name}
              margin="normal"
              helperText={!validateName(values.name) ? "Should be between 1 and 50 characters" : undefined}
              error={!validateName(values.name)}
              fullWidth
              required
              onChange={handleChange("name")}
            />
            <TextField
              id="bio"
              label="Bio"
              value={values.bio}
              margin="normal"
              helperText={!validateBio(values.bio) ? "Should be less than 255 characters" : undefined}
              error={!validateBio(values.bio)}
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
                loading || !validateUsername(values.username) || !validateName(values.name) || !validateBio(values.bio)
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

const validateUsername = (val) => {
  return val && val.length >= 3 && val.length <= 40 && val.match(/^[_a-z][_a-z0-9]+$/);
};

const validateName = (val) => {
  return val && val.length >= 1 && val.length <= 50;
};

const validateBio = (val) => {
  return val.length < 256;
};

const isUsernameError = (error) => {
  return error && error.message.match(/duplicate key value violates unique constraint/);
};
