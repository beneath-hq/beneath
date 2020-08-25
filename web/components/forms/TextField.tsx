import { InputBaseComponentProps, InputBase, makeStyles, Theme } from "@material-ui/core";
import { FC } from "react";

import FormControl, { FormControlProps } from "./FormControl";

const useStyles = makeStyles((theme: Theme) => ({
  root: {
    marginTop: "0.3rem",
    borderRadius: "4px",
    position: "relative",
    backgroundColor: theme.palette.background.paper,
    border: "1px solid",
    borderColor: "rgba(35, 48, 70, 1)",
    padding: "0 12px",
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
}));

export interface TextFieldProps extends FormControlProps {
  value?: string;
  multiline?: boolean;
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

const TextField: FC<TextFieldProps> = (props) => {
  const {
    id,
    value,
    multiline,
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
  const classes = useStyles();

  return (
    <FormControl id={id} {...others}>
      <InputBase
        classes={{
          root: classes.root,
          focused: classes.focused,
          input: classes.input,
          inputMultiline: classes.input,
        }}
        id={id}
        value={value}
        multiline={multiline}
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
