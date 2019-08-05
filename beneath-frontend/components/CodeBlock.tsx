import React, { PureComponent } from "react";

import { Theme, withTheme } from "@material-ui/core/styles";
import Highlight, { defaultProps, Language } from "prism-react-renderer";
import vsDark from "prism-react-renderer/themes/vsDark";

// Theme candidates:
//   dracula
//   nightOwl
//   oceanicNext
//   shadesOfPurple
//   vsDark

interface CodeBlockProps {
  theme: Theme;
  language: Language;
  children: string;
}

class CodeBlock extends PureComponent<CodeBlockProps> {
  public render() {
    const { language, theme, children } = this.props;

    const prismTheme = Object.assign({}, vsDark, {
      plain: {
        width: "100%",
        padding: "5px",
        overflowX: "auto",
        fontSize: "0.875rem",
        fontStyle: "normal",
        color: theme.palette.text,
        backgroundColor: theme.palette.background.paper,
      },
    });

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
  }
}

export default withTheme(CodeBlock);
