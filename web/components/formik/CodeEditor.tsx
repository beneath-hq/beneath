// Adapted from https://github.com/stackworx/formik-material-ui
import React, { FC } from "react";
import CodeEditor, { CodeEditorProps } from "../forms/CodeEditor";
import { FieldProps, getIn } from "formik";

export interface FormikCodeEditorProps extends FieldProps, Omit<CodeEditorProps, "name" | "value" | "error"> {}

export function formikToComponent({
  field: { name, value },
  form: { touched, errors, setFieldValue, setFieldTouched },
  onBlur,
  onChange,
  id,
  ...props
}: FormikCodeEditorProps): CodeEditorProps {
  const fieldError = getIn(errors, name);
  const showError = getIn(touched, name) && !!fieldError;
  const actualID = id ?? name;

  return {
    id: actualID,
    value,
    error: showError,
    errorText: showError && fieldError,
    onChange: (value) => {
      setFieldValue(name, value);
      if (onChange) {
        onChange(value || "");
      }
    },
    onBlur: () => {
      setFieldTouched(name, true, true);
      if (onBlur) {
        onBlur();
      }
    },
    ...props,
  };
}

export const FormikCodeEditor: FC<FormikCodeEditorProps> = (props) => {
  return <CodeEditor {...formikToComponent(props)} />;
};

export default FormikCodeEditor;
