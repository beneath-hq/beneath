// Adapted from https://github.com/stackworx/formik-material-ui
import { FieldProps, getIn } from "formik";
import React, { FC } from "react";

import Checkbox, { CheckboxProps } from "../forms/Checkbox";

export interface FormikCheckboxProps
  extends FieldProps,
    Omit<CheckboxProps, "name" | "value" | "error" | "form" | "checked" | "defaultChecked" | "type"> {
  type?: string;
}

export function formikToComponent({
  disabled,
  field: { onChange: formikOnChange, onBlur: formikOnBlur, ...field },
  form: { isSubmitting, touched, errors },
  type,
  onBlur,
  onChange,
  id,
  ...props
}: FormikCheckboxProps): CheckboxProps {
  const fieldError = getIn(errors, field.name);
  const showError = getIn(touched, field.name) && !!fieldError;
  const actualID = id ?? field.name;

  const indeterminate = !Array.isArray(field.value) && field.value == null;

  if (process.env.NODE_ENV !== "production") {
    if (type !== "checkbox") {
      console.error(`property type=checkbox is missing from field ${field.name}, this can caused unexpected behaviour`);
    }
  }

  return {
    error: showError,
    errorText: showError && fieldError,
    disabled: disabled ?? isSubmitting,
    indeterminate,
    onChange: (e, checked) => {
      formikOnChange(e);
      if (onChange) {
        onChange(e, checked);
      }
    },
    onBlur: (e) => {
      formikOnBlur(e);
      if (onBlur) {
        onBlur(e);
      }
    },
    id: actualID,
    ...field,
    ...props,
  };
}

export const FormikCheckbox: FC<FormikCheckboxProps> = (props) => {
  return <Checkbox {...formikToComponent(props)} />;
};

export default FormikCheckbox;
