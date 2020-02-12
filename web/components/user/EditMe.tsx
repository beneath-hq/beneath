import { useMutation, useQuery } from "@apollo/react-hooks";
import { useRouter } from "next/router";
import React, { FC } from "react";
import Moment from "react-moment";

import { Button, makeStyles, TextField, Typography } from "@material-ui/core";

import { QUERY_ME, UPDATE_ME } from "../../apollo/queries/user";
import { Me, Me_me } from "../../apollo/types/Me";
import { UpdateMe, UpdateMeVariables } from "../../apollo/types/UpdateMe";
import Loading from "../Loading";
import VSpace from "../VSpace";

const useStyles = makeStyles((theme) => ({
  submitButton: {
    marginTop: theme.spacing(3),
  },
}));

const EditMe = () => {
  const { loading, error, data } = useQuery<Me>(QUERY_ME);

  if (loading) {
    return <Loading justify="center" />;
  }

  if (error || !data || !data.me) {
    return <p>Error: {JSON.stringify(error)}</p>;
  }

  const { me } = data;
  return <EditMeForm me={me} />;
};

export default EditMe;

const EditMeForm: FC<{ me: Me_me }> = ({ me }) => {
  const [values, setValues] = React.useState({
    username: me.user.username || "",
    email: me.email || "",
    name: me.user.name || "",
    bio: me.user.bio || "",
    photoURL: me.user.photoURL || "",
  });

  const router = useRouter();

  const [updateMe, { loading, error }] = useMutation<UpdateMe, UpdateMeVariables>(UPDATE_ME, {
    onCompleted: (data) => {
      if (data.updateMe) {
        const username = data.updateMe.user.username;
        if (username !== router.query.name) {
          const href = `/user?name=${username}&tab=edit`;
          const as = `/users/${username}/edit`;
          router.replace(href, as, { shallow: true });
        }
      }
    },
  });

  const handleChange = (name: string) => (event: any) => {
    setValues({ ...values, [name]: event.target.value });
  };

  const classes = useStyles();
  return (
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
  );
};

const validateUsername = (val: string) => {
  return val && val.length >= 3 && val.length <= 40 && val.match(/^[_a-z][_a-z0-9]+$/);
};

const validateName = (val: string) => {
  return val && val.length >= 1 && val.length <= 50;
};

const validateBio = (val: string) => {
  return val.length < 256;
};

const isUsernameError = (error: Error) => {
  return error && error.message.match(/duplicate key value violates unique constraint/);
};
