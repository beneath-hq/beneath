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
  field: { onBlur: fieldOnBlur, ...field },
  form: { isSubmitting, touched, errors },
  type,
  onBlur,
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
    onBlur: onBlur ?? ((e) => fieldOnBlur(e ?? field.name)),
    id: actualID,
    ...field,
    ...props,
  };
}

export const FormikCheckbox: FC<FormikCheckboxProps> = (props) => {
  return <Checkbox {...formikToComponent(props)} />;
};

export default FormikCheckbox;
