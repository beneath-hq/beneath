import clsx from "clsx";
import React, { FC } from "react";

import { Box, makeStyles, Paper, Theme, Typography } from "@material-ui/core";
import { InsertDriveFile } from "@material-ui/icons";
import CodeBlock, { CodeBlockProps } from "components/CodeBlock";

const useStyles = makeStyles((theme: Theme) => ({
  header: {
    borderBottom: `1px solid ${theme.palette.border.paper}`,
    backgroundColor: theme.palette.background.medium,
    padding: theme.spacing(1),
  },
  code: {
    width: "100%",
    overflowX: "auto",
    padding: theme.spacing(2),
  },
  fileIcon: {
    marginRight: theme.spacing(0.5),
  },
  paragraph: {
    marginBottom: 16,
  },
}));

export interface CodePaperProps extends CodeBlockProps {
  filename?: string;
  paragraph?: boolean;
}

export const CodePaper: FC<CodePaperProps> = ({ filename, paragraph, ...blockProps }) => {
  const classes = useStyles();
  return (
    <Paper variant="outlined" className={clsx(paragraph && classes.paragraph)}>
      {filename && (
        <Box display="flex" className={classes.header}>
          <InsertDriveFile fontSize="small" className={classes.fileIcon} />
          <Typography variant="body2" color="textSecondary">
            {filename}
          </Typography>
        </Box>
      )}
      <div className={classes.code}>
        <CodeBlock {...blockProps} />
      </div>
    </Paper>
  );
};

export default CodePaper;
