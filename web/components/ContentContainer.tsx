import { Button, Container, makeStyles, Paper, Typography } from "@material-ui/core";
import { FC } from "react";
import Loading from "./Loading";

export interface CallToAction {
  message?: string;
  label?: string;
  onClick?: () => void;
}

export type ContentContainerProps = {
  paper?: boolean;
  maxWidth?: 'xs' | 'sm' | 'md' | 'lg' | 'xl' | false;
  margin?: "none" | "normal";
  loading?: boolean;
  error?: string;
  callToAction?: CallToAction;
  children?: any;
};

const useStyles = makeStyles((theme) => ({
  paper: {
    width: "100%",
    overflowX: "auto",
  },
  message: {
    margin: "5rem 0",
  },
  marginNormal: {
    marginTop: "1.25rem",
    marginBottom: "1.25rem",
  },
}));

const ContentContainer: FC<ContentContainerProps> = (props) => {
  const { paper, maxWidth, margin, loading, error, callToAction, children } = props;
  const classes = useStyles();

  let innerEl = (
    <>
      {children}
      {loading && (
        <div className={classes.message}>
          <Loading justify="center" />
        </div>
      )}
      {error && (
        <div className={classes.message}>
          <Typography color="error" align="center">
            {error}
          </Typography>
        </div>
      )}
      {!error && callToAction?.message && (
        <div className={classes.message}>
          <Typography color="textSecondary" align="center">
            {callToAction.message}
          </Typography>
          {callToAction.label && (
            <Button variant="contained" onClick={callToAction.onClick}>
              callToAction.label
            </Button>
          )}
        </div>
      )}
    </>
  );

  if (paper) {
    innerEl = <Paper className={classes.paper} variant="outlined">{innerEl}</Paper>;
  }

  if (maxWidth) {
    innerEl = <Container maxWidth={maxWidth}>{innerEl}</Container>;
  }

  if (margin === "normal") {
    innerEl = <div className={classes.marginNormal}>{innerEl}</div>;
  }

  return innerEl;
};

export default ContentContainer;
