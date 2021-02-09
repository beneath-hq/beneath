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
    fontFamily: theme.typography.fontFamilyMonospaced,
    fontSize: theme.typography.body2.fontSize,
    lineHeight: theme.typography.body2.lineHeight,
    margin: 0,
  };

  if (!language) {
    return <pre style={styles}>{children.trim()}</pre>;
  }

  const prismTheme = Object.assign({}, vsDark, { plain: styles });
  return (
    <Highlight {...defaultProps} code={children.trim()} language={language} theme={prismTheme}>
      {({ className, style, tokens, getLineProps, getTokenProps }) => (
        <pre className={className} style={style}>
          {tokens.map((line, i) => (
            <div {...getLineProps({ line, key: i, style: { overflowX: "none" } })}>
              {line.map((token, key) => {
                const tokenProps = getTokenProps({ token, key });
                if (line.length === 1 && tokenProps.children === "") {
                  tokenProps.children = "\n";
                }
                return <span {...tokenProps} />;
              })}
            </div>
          ))}
        </pre>
      )}
    </Highlight>
  );
};

export default CodeBlock;
