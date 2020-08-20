import React, { FC } from "react";
import Moment from "react-moment";

import { Button, makeStyles, Theme, Typography } from "@material-ui/core";
import { Alert } from "@material-ui/lab";

import FormControl, { FormControlProps } from "./FormControl";

const useStyles = makeStyles((theme: Theme) => ({
  alert: {
    marginBottom: 24,
  },
  button: {
    width: "fit-content", // keep small even in full width form
  },
  timestamps: {
    marginTop: 24,
  },
}));

export interface SubmitControlProps extends FormControlProps {
  label?: string;
  errorAlert?: string;
  createdOn?: string;
  updatedOn?: string;
}

const SubmitControl: FC<SubmitControlProps> = (props) => {
  const { label, errorAlert, createdOn, updatedOn, ...others } = props;
  const classes = useStyles();
  return (
    <FormControl {...others}>
      {errorAlert && (
        <Alert className={classes.alert} severity="error">
          {errorAlert}
        </Alert>
      )}
      <Button className={classes.button} type="submit" variant="contained" color="primary">
        {label}
      </Button>
      {(createdOn || updatedOn) && (
        <Typography className={classes.timestamps} variant="subtitle1" color="textSecondary">
          {createdOn && (
            <>
              Created <Moment fromNow date={createdOn} />
            </>
          )}
          {updatedOn && (createdOn ? " and last updated " : "Last updated ")}
          {updatedOn && <Moment fromNow date={updatedOn} />}
        </Typography>
      )}
    </FormControl>
  );
};

export default SubmitControl;
