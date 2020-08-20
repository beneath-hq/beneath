import React, { FC } from "react";

import {
  Checkbox as MuiCheckbox,
  CheckboxProps as MuiCheckboxProps,
  FormControl,
  FormControlLabel,
  FormHelperText,
  makeStyles,
} from "@material-ui/core";

import FieldError from "./FieldError";

const useStyles = makeStyles((theme) => ({
  checkbox: {
    "&&:hover": {
      color: theme.palette.primary.main,
      backgroundColor: "transparent",
    },
  },
  helper: {
    ...theme.typography.body2,
  },
}));

export interface CheckboxProps extends MuiCheckboxProps {
  label?: string;
  helperText?: string;
  error?: boolean;
  errorText?: string;
  disabled?: boolean;
  margin?: "none" | "dense" | "normal";
  fullWidth?: boolean;
}

const Checkbox: FC<CheckboxProps> = (props) => {
  const { id, label, helperText, error, errorText, disabled, margin, fullWidth, ...other } = props;
  const classes = useStyles();
  const actualFullWidth = fullWidth === undefined ? true : fullWidth;
  const actualMargin = margin ?? "normal";
  return (
    <FormControl margin={actualMargin} error={error} fullWidth={actualFullWidth} disabled={disabled}>
      <FormControlLabel label={label} control={<MuiCheckbox className={classes.checkbox} id={id} {...other} />} />
      {helperText && <FormHelperText className={classes.helper}>{helperText}</FormHelperText>}
      <FieldError id={id} error={error} errorText={errorText} />
    </FormControl>
  );
};

export default Checkbox;
