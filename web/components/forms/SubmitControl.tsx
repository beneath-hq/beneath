import React, { FC } from "react";
import Moment from "react-moment";

import { Button, FormControl, makeStyles, Theme, Typography } from "@material-ui/core";
import { Alert } from "@material-ui/lab";

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

export interface SubmitControlProps {
  label?: string;
  errorAlert?: string;
  createdOn?: string;
  updatedOn?: string;
  disabled?: boolean;
  margin?: "none" | "dense" | "normal";
  fullWidth?: boolean;
}

const SubmitControl: FC<SubmitControlProps> = (props) => {
  const { label, errorAlert, createdOn, updatedOn, disabled, margin, fullWidth } = props;
  const actualMargin = margin ?? "normal";
  const actualFullWidth = fullWidth === undefined ? true : fullWidth;
  const classes = useStyles();
  return (
    <FormControl margin={actualMargin} error={!!errorAlert} fullWidth={actualFullWidth} disabled={disabled}>
      {errorAlert && (
        <Alert className={classes.alert} severity="error">
          {errorAlert}
        </Alert>
      )}
      <Button className={classes.button} type="submit" variant="contained" color="primary" disabled={disabled}>
        {label}
      </Button>
      {(createdOn || updatedOn) && (
        <Typography className={classes.timestamps} variant="subtitle1" color="textSecondary">
          {createdOn && <>Created <Moment fromNow date={createdOn} /></>}
          {updatedOn && (createdOn ? " and last updated " : "Last updated ")}
          {updatedOn && <Moment fromNow date={updatedOn} />}
        </Typography>
      )}
    </FormControl>
  );
};

export default SubmitControl;
