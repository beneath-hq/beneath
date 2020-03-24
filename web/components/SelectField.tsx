import { FC } from "react";

import {
  FormControl,
  FormHelperText,
  Input,
  InputLabel,
  MenuItem,
  Select,
} from "@material-ui/core";

export interface SelectFieldProps {
  id: string;
  options: Array<{ label: string, value: string }>;
  value?: unknown;
  label?: string;
  required?: boolean;
  helperText?: string;
  onChange?: (
    event: React.ChangeEvent<{ name?: string; value: unknown }>,
    child: React.ReactNode,
  ) => void;
}

const SelectField: FC<SelectFieldProps> = (props) => {
  return (
    <FormControl>
      {props.label && (
        <InputLabel htmlFor={props.id} required={props.required}>
          {props.label}
        </InputLabel>
      )}
      <Select value={props.value} onChange={props.onChange} input={<Input id={props.id} name={props.id} />}>
        {props.options.map((option) => (
          <MenuItem key={option.value} value={option.value}>
            {option.label}
          </MenuItem>
        ))}
      </Select>
      {props.helperText && <FormHelperText>{props.helperText}</FormHelperText>}
    </FormControl>
  );
};

export default SelectField;
