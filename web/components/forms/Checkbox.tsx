import React, { FC } from "react";

import {
  Checkbox as MuiCheckbox,
  CheckboxProps as MuiCheckboxProps,
  FormControlLabel,
  FormHelperText,
  makeStyles,
} from "@material-ui/core";

import FormControl, { FormControlProps } from "./FormControl";

const useStyles = makeStyles((theme) => ({
  root: {
    "&&:hover": {
      color: theme.palette.primary.main,
      backgroundColor: "transparent",
    },
    '&$checked': {
      color: theme.palette.primary.main,
    }
  },
  checked: {},
  helper: {
    ...theme.typography.body2,
  },
}));

export interface CheckboxProps extends FormControlProps, MuiCheckboxProps {}

const Checkbox: FC<CheckboxProps> = (props) => {
  const {
    id,
    label,
    margin,
    helperText,
    error,
    errorText,
    fullWidth,
    required, // ignored
    disabled,
    ...muiCheckboxProps
  } = props;

  const classes = useStyles();
  return (
    <FormControl
      id={id}
      margin={margin}
      helperText={helperText}
      error={error}
      errorText={errorText}
      fullWidth={fullWidth}
      disabled={disabled}
    >
      <FormControlLabel
        label={label}
        control={<MuiCheckbox classes={{root: classes.root, checked: classes.checked}} id={id} {...muiCheckboxProps} />}
      />
      {helperText && <FormHelperText className={classes.helper}>{helperText}</FormHelperText>}
    </FormControl>
  );
};

export default Checkbox;
