// Adapted from https://github.com/stackworx/formik-material-ui
import React, { FC } from "react";
import TextField, { TextFieldProps } from "../forms/TextField";
import { FieldProps, getIn } from "formik";

export interface FormikTextFieldProps extends FieldProps, Omit<TextFieldProps, "name" | "value" | "error"> {}

export function formikToComponent({
  field: { onChange: formikOnChange, onBlur: formikOnBlur, ...field },
  form: { isSubmitting, touched, errors },
  disabled,
  onBlur,
  onChange,
  id,
  ...props
}: FormikTextFieldProps): TextFieldProps {
  const fieldError = getIn(errors, field.name);
  const showError = getIn(touched, field.name) && !!fieldError;
  const actualID = id ?? field.name;

  return {
    error: showError,
    errorText: showError && fieldError,
    disabled: disabled ?? isSubmitting,
    onChange: ((...args) => {
      formikOnChange(...args);
      if (onChange) {
        onChange(...args);
      }
    }),
    onBlur: ((...args) => {
      formikOnBlur(...args);
      if (onBlur) {
        onBlur(...args);
      }
    }),
    id: actualID,
    ...field,
    ...props,
  };
}

export const FormikTextField: FC<FormikTextFieldProps> = (props) => {
  return <TextField {...formikToComponent(props)} />;
};

export default FormikTextField;
