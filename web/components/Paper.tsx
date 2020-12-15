import { Box, BoxProps, Grid as MuiGrid, GridProps as MuiGridProps, makeStyles, Paper as MuiPaper, PaperProps as MuiPaperProps, Theme } from "@material-ui/core";
import { FC } from "react";

export interface PaperProps extends MuiPaperProps {
  padding?: "dense" | "normal" | "large";
}

export const Paper: FC<PaperProps> = ({ padding, ...props }) => {
  const space = padding === "dense" ? 1.5 : padding === "large" ? 4 : 2.5;
  const box: FC<BoxProps> = (boxProps) => <Box p={space} {...boxProps} />;
  return (
    <MuiPaper component={box} {...props} />
  );
};

export interface PaperGridProps extends MuiGridProps {
  padding?: "dense" | "normal" | "large";
  variant?: "elevation" | "outlined";
}

export const PaperGrid: FC<PaperGridProps> = ({ padding, variant, children, className, ...props }) => {
  if (props.item && props.container) {
    console.warn("PaperGrid: don't set 'item' and 'container' on the same component")
  } else if (props.item) {
    return (
      <MuiGrid {...props}>
        <Paper padding={padding} variant={variant} className={className}>
          {children}
        </Paper>
      </MuiGrid>
    );
  } else if (props.container) {
    return (
      <Paper padding={padding} variant={variant} className={className}>
        <MuiGrid {...props}>{children}</MuiGrid>
      </Paper>
    );
  }

  console.warn("PaperGrid: neither 'item' nor 'container' set");
  return <></>;
};
