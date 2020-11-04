import React, { FC } from "react";
import clsx from "clsx";

import { makeStyles, Theme} from "@material-ui/core";
import Highlight, { defaultProps, Language } from "prism-react-renderer";
import vsDark from "prism-react-renderer/themes/vsDark";
import { InsertDriveFile } from "@material-ui/icons";

// Theme candidates:
//   dracula
//   nightOwl
//   oceanicNext
//   shadesOfPurple
//   vsDark

const useStyles = makeStyles((theme: Theme) => ({
  block: {
    backgroundColor: theme.palette.background.paper,
    overflowX: "auto",
  },
  blockNoHeader: {
    border: `1px solid ${theme.palette.border.paper}`,
    borderRadius: "4px",
  },
  blockHeader: {
    borderLeft: `1px solid ${theme.palette.border.paper}`,
    borderRight: `1px solid ${theme.palette.border.paper}`,
    borderBottom: `1px solid ${theme.palette.border.paper}`,
    borderBottomLeftRadius: "4px",
    borderBottomRightRadius: "4px",
  },
  header: {
    paddingLeft: theme.spacing(.5),
    paddingTop: theme.spacing(.5),
    backgroundColor: theme.palette.background.default,
    border: `1px solid ${theme.palette.border.paper}`,
    borderRadius: "4px 4px 0px 0px",
  },
  headerContent: {
    display: "inline-flex",
    alignItems: "center",
    opacity: .5
  },
  headerIcon: {
    height: "20px",
    marginRight: theme.spacing(.5),
  },
  pre: {
    fontFamily: theme.typography.fontFamilyMonospaced,
    fontSize: theme.typography.body2.fontSize,
    textAlign: "left",
  },
  tokenLine: {
    display: "table-row",
  },
  token: {
  },
  lineNo: {
    display: "table-cell",
    textAlign: "right",
    paddingRight: theme.spacing(3),
    paddingLeft: theme.spacing(1.5),
    userSelect: "none",
    opacity: .5,
  },
  lineContent: {
    display: "table-cell",
  }
}));

export interface CodeBlockProps {
  language?: Language;
  title?: string;
  children: string;
}

export const CodeBlock: FC<CodeBlockProps> = ({ language, title, children }) => {
  const classes = useStyles();
  const styles: React.CSSProperties = {};

  if (!language) {
    return <pre style={styles}>{children}</pre>;
  }

  const prismTheme = Object.assign({}, vsDark, { plain: styles });
  return (
    <>
      {title && (
        <div className={classes.header}>
          <div className={classes.headerContent}>
            <InsertDriveFile className={classes.headerIcon} />
            {title}
          </div>
        </div>
      )}
      <div className={clsx(classes.block, !title && classes.blockNoHeader, title && classes.blockHeader)}>
        <Highlight {...defaultProps} code={children} language={language} theme={prismTheme}>
          {({ className, style, tokens, getLineProps, getTokenProps }) => (
            <pre className={classes.pre}>
              {tokens.map((line, i) => (
                <div key={i} className={classes.tokenLine}>
                  <span className={classes.lineNo}>{i + 1}</span>
                  <span className={classes.lineContent}>
                    {line.map((token, key) => (
                      <span key={key} {...getTokenProps({ token, key })} />
                    ))}
                  </span>
                </div>
              ))}
            </pre>
          )}
        </Highlight>
      </div>
    </>
  );
};

export default CodeBlock;
