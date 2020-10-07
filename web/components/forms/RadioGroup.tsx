import { FormControlLabel, makeStyles, Radio, RadioGroup as MuiRadioGroup, Theme } from "@material-ui/core";
import React from "react";

import FormControl, { FormControlProps } from "./FormControl";

export interface RadioOption {
  value: string;
  label: string;
}

export interface RadioGroupProps<T> extends FormControlProps {
  options: RadioOption[];
  value: string;
  name?: string;
  row?: boolean;
  onBlur?: React.FocusEventHandler<HTMLInputElement>;
  onChange?: (event: React.ChangeEvent<HTMLInputElement>, value: string) => void;
}

const useStyles = makeStyles((theme: Theme) => ({}));

function RadioGroup<T>(props: RadioGroupProps<T>) {
  const { options, value, name, row, onBlur, onChange, ...formControlProps } = props;
  const classes = useStyles();
  return (
    <FormControl {...formControlProps}>
      <MuiRadioGroup value={value} name={name} row={row} onChange={onChange} onBlur={onBlur}>
        {options.map((option, idx) => (
          <FormControlLabel
            key={idx}
            value={option.value}
            control={<Radio color="default" />}
            label={option.label}
          />
        ))}
      </MuiRadioGroup>
    </FormControl>
  );
}

export default RadioGroup;
