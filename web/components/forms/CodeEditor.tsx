import { FC } from "react";

import FormControl, { FormControlProps } from "./FormControl";
import ActualCodeEditor, { CodeEditorProps as ActualCodeEditorProps } from "../CodeEditor";

export interface CodeEditorProps extends ActualCodeEditorProps, Omit<FormControlProps, "fullWidth" | "disabled" | "children"> {}

const CodeEditor: FC<CodeEditorProps> = (props) => {
  const {
    rows,
    language,
    value,
    onChange,
    onFocus,
    onBlur,
    id,
    margin,
    ...others
  } = props;
  return (
    <FormControl id={id} margin={margin} fullWidth {...others}>
      <ActualCodeEditor
        rows={rows}
        language={language}
        value={value}
        onChange={onChange}
        onFocus={onFocus}
        onBlur={onBlur}
      />
    </FormControl>
  );
};

export default CodeEditor;
