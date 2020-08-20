import React, { FC } from "react";
import TextField, { TextFieldProps } from "../forms/TextField";
import { FieldProps, getIn } from "formik";

export interface FormikTextFieldProps extends FieldProps, Omit<TextFieldProps, "name" | "value" | "error"> {}

export function formikToComponent({
  field: { onBlur: fieldOnBlur, ...field },
  form: { isSubmitting, touched, errors },
  disabled,
  onBlur,
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
    onBlur: onBlur ?? ((e) => fieldOnBlur(e ?? field.name)),
    id: actualID,
    ...field,
    ...props,
  };
}

export const FormikTextField: FC<FormikTextFieldProps> = (props) => {
  return <TextField {...formikToComponent(props)} />;
};

export default FormikTextField;
