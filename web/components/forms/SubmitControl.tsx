import { Button, Grid, makeStyles, Theme, Typography } from "@material-ui/core";
import { Alert } from "@material-ui/lab";
import React, { FC } from "react";
import Moment from "react-moment";

import clsx from "clsx";
import FormControl, { FormControlProps } from "./FormControl";

const useStyles = makeStyles((theme: Theme) => ({
  alert: {
    marginBottom: 24,
  },
  button: {
    width: "fit-content", // keep small even in full width form
  },
  severeButton: {
    backgroundColor: theme.palette.error.main,
  },
  timestamps: {
    marginTop: 24,
  },
}));

export interface SubmitControlProps extends FormControlProps {
  label?: string;
  cancelFn?: () => void;
  cancelLabel?: string;
  severe?: boolean;
  errorAlert?: string;
  createdOn?: string;
  updatedOn?: string;
  rightSide?: boolean;
}

const SubmitControl: FC<SubmitControlProps> = (props) => {
  const { label, cancelFn, cancelLabel, severe, errorAlert, createdOn, updatedOn, disabled, rightSide, ...others } = props;
  const classes = useStyles();
  return (
    <FormControl disabled={disabled} {...others}>
      {errorAlert && (
        <Alert className={classes.alert} severity="error">
          {errorAlert}
        </Alert>
      )}
      <Grid container alignItems="center" spacing={2} justify={rightSide ? "flex-end" : "flex-start"}>
        {cancelFn && cancelLabel && (
          <Grid item>
            <Button className={classes.button} onClick={cancelFn}>
              {cancelLabel}
            </Button>
          </Grid>
        )}
        <Grid item>
          <Button className={clsx(classes.button, severe && classes.severeButton)} type="submit" disabled={disabled} variant="contained" color="primary">
            {label}
          </Button>
        </Grid>
      </Grid>
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
