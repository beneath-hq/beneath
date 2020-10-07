import React, { FC } from "react";

import { useTheme } from "@material-ui/core/styles";
import Highlight, { defaultProps, Language } from "prism-react-renderer";
import vsDark from "prism-react-renderer/themes/vsDark";

// Theme candidates:
//   dracula
//   nightOwl
//   oceanicNext
//   shadesOfPurple
//   vsDark

export interface CodeBlockProps {
  language?: Language;
  children: string;
}

export const CodeBlock: FC<CodeBlockProps> = ({ language, children }) => {
  const theme = useTheme();
  const styles: React.CSSProperties = {
    // backgroundColor: theme.palette.background.paper,
    // padding: theme.spacing(0.75),
    fontFamily: theme.typography.fontFamilyMonospaced,
    fontSize: theme.typography.body2.fontSize,
    overflowX: "auto",
    width: "100%",
  };

  if (!language) {
    return <pre style={styles}>{children}</pre>;
  }

  const prismTheme = Object.assign({}, vsDark, { plain: styles });
  return (
    <Highlight {...defaultProps} code={children} language={language} theme={prismTheme}>
      {({ className, style, tokens, getLineProps, getTokenProps }) => (
        <pre className={className} style={style}>
          {tokens.map((line, i) => (
            <div {...getLineProps({ line, key: i })}>
              {line.map((token, key) => (
                <span {...getTokenProps({ token, key })} />
              ))}
            </div>
          ))}
        </pre>
      )}
    </Highlight>
  );
};

export default CodeBlock;
