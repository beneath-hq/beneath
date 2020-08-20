import { FormControl, InputBase, makeStyles, Theme } from "@material-ui/core";
import { FC } from "react";

import FieldLabel from "./FieldLabel";
import FieldError from "./FieldError";

const useStyles = makeStyles((theme: Theme) => ({
  input: {
    borderRadius: "4px",
    position: "relative",
    backgroundColor: theme.palette.background.paper,
    border: "1px solid",
    borderColor: "rgba(35, 48, 70, 1)",
    padding: "10px 12px",
    marginTop: "0.3rem",
    "&:focus": {
      boxShadow: `0 0 0 2px ${theme.palette.primary.main}`,
    },
  },
  multiline: {
    padding: "0",
  },
}));

export type TextFieldProps = {
  id?: string;
  label?: string;
  value?: string;
  margin?: "none" | "dense" | "normal";
  helperText?: string;
  error?: boolean;
  errorText?: string;
  fullWidth?: boolean;
  multiline?: boolean;
  rows?: number;
  rowsMax?: number;
  required?: boolean;
  disabled?: boolean;
  onChange?: React.ChangeEventHandler<HTMLTextAreaElement | HTMLInputElement>;
  onBlur?: React.FocusEventHandler<HTMLInputElement | HTMLTextAreaElement>;
};

const TextField: FC<TextFieldProps> = (props) => {
  const {
    id,
    label,
    value,
    margin,
    helperText,
    error,
    errorText,
    fullWidth,
    multiline,
    rows,
    rowsMax,
    required,
    onChange,
    onBlur,
    disabled,
  } = props;
  const classes = useStyles();
  const actualFullWidth = fullWidth === undefined ? true : fullWidth;
  const actualMargin = margin ?? "normal";
  return (
    <FormControl margin={actualMargin} error={error} fullWidth={actualFullWidth} disabled={disabled}>
      <FieldLabel id={id} label={label} helperText={helperText} required={required} />
      <InputBase
        classes={{ input: classes.input, multiline: classes.multiline }}
        id={id}
        value={value}
        multiline={multiline}
        rows={rows}
        rowsMax={rowsMax}
        onChange={onChange}
        onBlur={onBlur}
        fullWidth={actualFullWidth}
      />
      <FieldError id={id} error={error} errorText={errorText} />
    </FormControl>
  );
};

export default TextField;
