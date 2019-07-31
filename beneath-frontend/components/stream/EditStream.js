import React from "react";
import Moment from "react-moment";
import { Mutation } from "react-apollo";

import Button from "@material-ui/core/Button";
import FormGroup from "@material-ui/core/FormGroup";
import TextField from "@material-ui/core/TextField";
import Typography from "@material-ui/core/Typography";
import { makeStyles } from "@material-ui/core";

import VSpace from "../VSpace";
import CheckboxField from "../CheckboxField";

import { UPDATE_STREAM } from "../../apollo/queries/stream";

const useStyles = makeStyles((theme) => ({
  submitButton: {
    marginTop: theme.spacing(3),
  },
}));

const EditStream = ({ stream }) => {
  const [values, setValues] = React.useState({
    description: stream.description || "",
    manual: stream.manual,
  });

  const handleChange = (name) => (event) => {
    setValues({ ...values, [name]: event.target.value });
  };

  const handleCheckboxChange = (name) => (event) => {
    setValues({ ...values, [name]: event.target.checked });
  };

  const classes = useStyles();
  return (
    <Mutation mutation={UPDATE_STREAM}>
      {(updateStream, { loading, error }) => (
        <div>
          <form onSubmit={(e) => {
            e.preventDefault();
            updateStream({ variables: { streamID: stream.streamID, ...values} });
          }}
          >
            <TextField
              id="description"
              label="Description"
              value={values.description}
              margin="normal"
              fullWidth
              required
              onChange={handleChange("description")}
              helperText="Help people understand what data this stream contains"
            />
            <FormGroup>
              <CheckboxField
                label="Allow manual entry"
                checked={values.manual}
                onChange={handleCheckboxChange("manual")}
                margin="normal"
                helperText="Enable writing data to the stream through the UI"
              />
            </FormGroup>
            <Button type="submit" variant="outlined" color="primary" className={classes.submitButton}
              disabled={
                loading
                || !(values.description && values.description.length <= 255)
              }>
              Save changes
            </Button>
            {error && (
              <Typography variant="body1" color="error">An error occurred</Typography>
            )}
          </form>
          <VSpace units={2} />
          <Typography variant="subtitle1" color="textSecondary">
            Stream created <Moment fromNow date={stream.createdOn} /> and
            its details last updated <Moment fromNow date={stream.updatedOn} />.
          </Typography>
        </div>
      )}
    </Mutation>
  );
};

export default EditStream;
