import { makeStyles, Theme } from "@material-ui/core";
import {
  Autocomplete as MuiAutocomplete,
  AutocompleteChangeReason,
  AutocompleteChangeDetails,
  AutocompleteInputChangeReason,
} from "@material-ui/lab";
import { FocusEventHandler } from "react";

import FormControl, { FormControlProps } from "./FormControl";
import { InputField } from "./TextField";

export interface SelectFieldProps<T> extends FormControlProps {
  options: T[];
  getOptionLabel: (option: T) => string;
  value: T | T[] | null;
  loading?: boolean;
  multiple?: boolean;
  showClearButton?: boolean;
  disableClearable?: boolean;
  onBlur?: FocusEventHandler<HTMLDivElement>;
  onChange?: (
    event: React.ChangeEvent<{}>,
    value: T | NonNullable<T> | T[] | null,
    reason: AutocompleteChangeReason,
    details?: AutocompleteChangeDetails<T>
  ) => void;
  onInputChange?: (event: React.ChangeEvent<{}>, value: string, reason: AutocompleteInputChangeReason) => void;
  getOptionSelected?: (option: T, value: T) => boolean;
}

const useStyles = makeStyles((theme: Theme) => ({
  endAdornment: {
    top: "unset",
    marginRight: "12px",
  },
  hideClearIndicator: {
    display: "none",
  },
}));

function SelectField<T>(props: SelectFieldProps<T>) {
  const {
    id,
    value,
    disabled,
    fullWidth,
    options,
    getOptionLabel,
    getOptionSelected,
    loading,
    multiple,
    disableClearable,
    showClearButton,
    onBlur,
    onChange,
    onInputChange,
    ...formControlProps
  } = props;
  const classes = useStyles();
  return (
    <FormControl id={id} disabled={disabled} fullWidth={fullWidth} {...formControlProps}>
      <MuiAutocomplete<T, boolean, boolean, undefined>
        id={id}
        value={value}
        disabled={disabled}
        fullWidth={true}
        options={options}
        getOptionLabel={getOptionLabel}
        getOptionSelected={getOptionSelected}
        loading={loading}
        multiple={multiple}
        onBlur={onBlur}
        onChange={onChange}
        onInputChange={onInputChange}
        classes={{
          endAdornment: classes.endAdornment,
          clearIndicator: showClearButton ? undefined : classes.hideClearIndicator,
        }}
        disableClearable={disableClearable}
        openOnFocus
        renderInput={(params) => {
          const { id, inputProps, InputProps } = params;
          return (
            <InputField
              id={id}
              inputRef={InputProps.ref}
              inputProps={inputProps}
              startAdornment={InputProps.startAdornment}
              endAdornment={InputProps.endAdornment}
            />
          );
        }}
      />
    </FormControl>
  );
}

export default SelectField;
