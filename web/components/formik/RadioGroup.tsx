// Adapted from https://github.com/stackworx/formik-material-ui
import { FieldProps, getIn } from "formik";
import React, { FC } from "react";

import RadioGroup, { RadioGroupProps } from "../forms/RadioGroup";

export interface FormikRadioGroupProps<T> extends FieldProps, Omit<RadioGroupProps<T>, "name" | "value"> {}

export function formikToComponent<T>({
  field: { onBlur: formikOnBlur, ...field },
  form: { isSubmitting, touched, errors },
  onBlur,
  disabled,
  id,
  ...props
}: FormikRadioGroupProps<T>): RadioGroupProps<T> {
  const fieldError = getIn(errors, field.name);
  const showError = getIn(touched, field.name) && !!fieldError;
  const actualID = id ?? field.name;
  return {
    error: showError,
    errorText: showError && fieldError,
    disabled: disabled ?? isSubmitting,
    onBlur: (e) => {
      formikOnBlur(e);
    },
    id: actualID,
    ...field,
    ...props,
  };
}

export function FormikRadioGroup<T>(props: FormikRadioGroupProps<T>) {
  return <RadioGroup {...formikToComponent(props)} />;
}

export default FormikRadioGroup;
