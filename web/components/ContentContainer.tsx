import { Box, Button, Container, Grid, makeStyles, Paper, Typography } from "@material-ui/core";
import { FC } from "react";
import { NakedLink } from "./Link";
import Loading from "./Loading";

export interface CallToAction {
  message?: string | JSX.Element;
  buttons?: { label: string, href: string, as?: string }[];
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
        <div className={classes.message} >
        <Grid container spacing={4}>
          <Grid item xs={12}>
            <Typography color="textSecondary" align="center">
              {callToAction.message}
            </Typography>
          </Grid>
          <Grid item xs={12}>
            <Grid container justify="center" spacing={2}>
              {callToAction.buttons?.map((cta, idx) => (
                <Grid item key={idx}>
                  <Button variant="contained" component={NakedLink} href={cta.href} as={cta.as}>
                    {cta.label}
                  </Button>
                </Grid>
              )) }
            </Grid>
          </Grid>
        </Grid>
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
