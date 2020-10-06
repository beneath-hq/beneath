import { Box, FormControl as MuiFormControl } from "@material-ui/core";
import { FC } from "react";

import FieldLabel from "./FieldLabel";
import FieldError from "./FieldError";

export type FormControlProps = {
  id?: string;
  label?: React.ReactNode;
  margin?: "none" | "dense" | "normal";
  helperText?: string;
  error?: boolean;
  errorText?: string;
  fullWidth?: boolean;
  required?: boolean;
  disabled?: boolean;
  children?: any;
};

const FormControl: FC<FormControlProps> = (props) => {
  const {
    id,
    label,
    margin,
    helperText,
    error,
    errorText,
    fullWidth,
    required,
    disabled,
    children,
  } = props;
  const actualFullWidth = fullWidth === undefined ? true : fullWidth;
  const actualMargin = margin ?? "normal";
  return (
    <MuiFormControl margin={actualMargin} error={error} fullWidth={actualFullWidth} disabled={disabled}>
      {id && label && <FieldLabel id={id} label={label} helperText={helperText} required={required} />}
      <Box height="0.3rem" />
      {children}
      {id && <FieldError id={id} error={error} errorText={errorText} />}
    </MuiFormControl>
  );
};

export default FormControl;
