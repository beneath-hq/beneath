import { useTheme, lighten, rgbToHex, makeStyles, Box } from "@material-ui/core";
import clsx from "clsx";
import { FC, useState } from "react";
import Editor from "@monaco-editor/react";

const useStyles = makeStyles((theme) => ({
  container: {
    backgroundColor: theme.palette.background.paper,
    border: `1px solid ${theme.palette.border.paper}`,
    borderRadius: "4px",
    overflow: "hidden",
  },
  containerFocus: {
    boxShadow: `0 0 0 2px ${theme.palette.primary.main}`,
  },
}));

export interface CodeEditorProps {
  rows: number;
  language?: string;
  value: string | null;
  onChange: (value: string | undefined) => void;
  onFocus?: () => void;
  onBlur?: () => void;
}

const CodeEditor: FC<CodeEditorProps> = (props) => {
  const { rows, language, value, onChange, onFocus, onBlur } = props;
  const height = rows * 25;

  const [focus, setFocus] = useState(false);
  const theme = useTheme();
  const classes = useStyles();
  return (
    <Box className={clsx(classes.container, focus && classes.containerFocus)} height={height}>
      <Editor
        height={height}
        language={language}
        value={value || undefined}
        onChange={onChange}
        theme="beneath-theme"
        options={{
          automaticLayout: true,
          contextmenu: false,
          folding: false,
          fontFamily: theme.typography.fontFamilyMonospaced,
          fontSize: 14,
          lineDecorationsWidth: 20,
          lineNumbersMinChars: 4,
          minimap: {
            enabled: false,
          },
          // mouseWheelZoom: true,
          padding: {
            top: 15,
            bottom: 15,
          },
          scrollBeyondLastLine: false,
          selectOnLineNumbers: true,
        }}
        beforeMount={(monaco) => {
          monaco.editor.defineTheme("beneath-theme", {
            base: "vs-dark",
            inherit: true,
            rules: [],
            colors: {
              "editor.background": theme.palette.background.paper,
              "editor.lineHighlightBackground": theme.palette.background.medium,
              "editorLineNumber.foreground": rgbToHex(lighten(theme.palette.background.paper, 0.3)),
            },
          });
        }}
        onMount={(editor, monaco) => {
          editor.onDidFocusEditorText(() => {
            if (onFocus) onFocus();
            setFocus(true);
          });
          editor.onDidBlurEditorText(() => {
            if (onBlur) onBlur();
            setFocus(false);
          });
        }}
      />
    </Box>
  );
};

export default CodeEditor;
