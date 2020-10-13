import { InputBaseComponentProps, InputBase, makeStyles, Paper, Theme } from "@material-ui/core";
import { FC } from "react";

import FormControl, { FormControlProps } from "./FormControl";
import clsx from "clsx";

const useStyles = makeStyles((theme: Theme) => ({
  root: {
    borderRadius: "4px",
    position: "relative",
    padding: "0 12px",
    // outlined paper styles
    backgroundColor: theme.palette.background.paper,
    border: `1px solid ${theme.palette.border.paper}`,
  },
  focused: {
    boxShadow: `0 0 0 2px ${theme.palette.primary.main}`,
  },
  input: {
    padding: "10px 0",
    height: "100%",
  },
  multiline: {
    padding: "0",
  },
  monospace: {
    fontFamily: theme.typography.fontFamilyMonospaced,
    fontSize: theme.typography.body2.fontSize,
  },
}));

export interface InputFieldProps {
  id?: string;
  value?: string;
  placeholder?: string;
  multiline?: boolean;
  monospace?: boolean;
  rows?: number;
  rowsMax?: number;
  inputProps?: InputBaseComponentProps;
  inputRef?: React.Ref<any>;
  onBlur?: React.FocusEventHandler<HTMLInputElement | HTMLTextAreaElement>;
  onChange?: React.ChangeEventHandler<HTMLInputElement | HTMLTextAreaElement>;
  onFocus?: React.FocusEventHandler<HTMLInputElement | HTMLTextAreaElement>;
  endAdornment?: React.ReactNode;
  startAdornment?: React.ReactNode;
}

export const InputField: FC<InputFieldProps> = (props) => {
  const {
    id,
    value,
    placeholder,
    multiline,
    monospace,
    rows,
    rowsMax,
    inputProps,
    inputRef,
    onBlur,
    onChange,
    onFocus,
    endAdornment,
    startAdornment,
  } = props;
  const classes = useStyles();
  return (
    <InputBase
      classes={{
        root: clsx(classes.root, monospace && classes.monospace),
        focused: classes.focused,
        input: classes.input,
        inputMultiline: classes.input,
      }}
      id={id}
      value={value}
      placeholder={placeholder}
      multiline={multiline}
      fullWidth={true}
      rows={rows}
      rowsMax={rowsMax}
      inputProps={inputProps}
      inputRef={inputRef}
      onBlur={onBlur}
      onChange={onChange}
      onFocus={onFocus}
      endAdornment={endAdornment}
      startAdornment={startAdornment}
      spellCheck={multiline && !monospace}
    />
  );
};

export type TextFieldProps = InputFieldProps & FormControlProps;

export const TextField: FC<TextFieldProps> = (props) => {
  const {
    id,
    value,
    placeholder,
    multiline,
    monospace,
    rows,
    rowsMax,
    inputProps,
    inputRef,
    onBlur,
    onChange,
    onFocus,
    endAdornment,
    startAdornment,
    ...others
  } = props;

  return (
    <FormControl id={id} {...others}>
      <InputField
        id={id}
        value={value}
        placeholder={placeholder}
        multiline={multiline}
        monospace={monospace}
        rows={rows}
        rowsMax={rowsMax}
        inputProps={inputProps}
        inputRef={inputRef}
        onBlur={onBlur}
        onChange={onChange}
        onFocus={onFocus}
        endAdornment={endAdornment}
        startAdornment={startAdornment}
      />
    </FormControl>
  );
};

export default TextField;
