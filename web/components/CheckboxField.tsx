import React, { FC } from "react";

import {
  Checkbox,
  FormControl,
  FormControlLabel,
  FormHelperText,
  makeStyles,
  PropTypes,
} from "@material-ui/core";

const useStyles = makeStyles((theme) => ({
  checkbox: {
    "&&:hover": {
      backgroundColor: "transparent",
    },
  },
}));

export interface CheckboxFieldProps {
  label: React.ReactNode;
  checked?: boolean;
  helperText?: string;
  onChange?: (event: React.ChangeEvent<HTMLInputElement>, checked: boolean) => void;
  fullWidth?: boolean;
  margin?: PropTypes.Margin;
}

const CheckboxField: FC<CheckboxFieldProps> = ({ label, checked, onChange, helperText, ...other }) => {
  const classes = useStyles();
  return (
    <FormControl {...other}>
      <FormControlLabel
        label={label}
        control={<Checkbox className={classes.checkbox} color="primary" checked={checked} onChange={onChange} />}
      />
      {helperText && <FormHelperText>{helperText}</FormHelperText>}
    </FormControl>
  );
};

export default CheckboxField;
