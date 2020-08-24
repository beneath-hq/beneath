// Adapted from https://github.com/stackworx/formik-material-ui
import SelectField, { SelectFieldProps } from "../forms/SelectField";

import * as React from "react";
import { FieldProps, getIn } from "formik";

export interface FormikSelectFieldProps<T>
  extends FieldProps, Omit<SelectFieldProps<T>, "name" | "value" | "defaultValue"> {
  type?: string;
}

export function formikToComponent<T>({
  disabled,
  field,
  form: { isSubmitting, setFieldValue, touched, errors },
  type,
  id,
  onChange,
  onBlur,
  onInputChange,
  ...props
}: FormikSelectFieldProps<T>): SelectFieldProps<T> {
  if (process.env.NODE_ENV !== "production") {
    if (props.multiple) {
      if (!Array.isArray(field.value)) {
        console.error(`value for ${field.name} is not an array, this can caused unexpected behaviour`);
      }
    }
  }

  const fieldError = getIn(errors, field.name);
  const showError = getIn(touched, field.name) && !!fieldError;
  const actualID = id ?? field.name;

  const { onChange: formikOnChange, onBlur: formikOnBlur, multiple: _multiple, ...fieldSubselection } = field;

  // NOTE: If we want to enable freeSolo in the future, see the original Autocomplete example
  // from formik-material-ui (requires some non-trivial callback logic to work correctly).

  return {
    id: actualID,
    error: showError,
    errorText: showError && fieldError,
    onChange: (...args) => {
      const [_, value] = args;
      setFieldValue(field.name, value);
      if (onChange) {
        onChange(...args);
      }
    },
    onBlur: (...args) => {
      formikOnBlur(...args);
      if (onBlur) {
        onBlur(...args);
      }
    },
    onInputChange,
    disabled: disabled ?? isSubmitting,
    loading: isSubmitting,
    ...fieldSubselection,
    ...props,
  };
}

export function FormikSelectField<T>(props: FormikSelectFieldProps<T>) {
  return <SelectField {...formikToComponent(props)} />;
}

export default FormikSelectField;
