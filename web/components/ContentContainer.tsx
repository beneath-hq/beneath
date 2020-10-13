import { Box, Button, Container, Grid, makeStyles, Paper, Typography } from "@material-ui/core";
import { FC } from "react";
import { NakedLink } from "./Link";
import Loading from "./Loading";

export interface CallToAction {
  message?: string | JSX.Element;
  buttons?: { label: string; onClick?: () => void; href?: string; as?: string }[];
}

export type ContentContainerProps = {
  paper?: boolean;
  maxWidth?: "xs" | "sm" | "md" | "lg" | "xl" | false;
  margin?: "none" | "normal";
  loading?: boolean;
  error?: string;
  note?: string | JSX.Element;
  callToAction?: CallToAction;
  children?: any;
};

const useStyles = makeStyles((theme) => ({
  paper: {
    width: "100%",
    overflowX: "auto",
  },
  content: {
    margin: "3.5rem 0",
  },
  note: {
    marginTop: theme.spacing(2),
  },
  marginNormal: {
    marginTop: "1.25rem",
    marginBottom: "1.25rem",
  },
}));

const ContentContainer: FC<ContentContainerProps> = (props) => {
  const { paper, maxWidth, margin, loading, error, note, callToAction, children } = props;
  const classes = useStyles();

  let innerEl = (
    <>
      {children}
      {loading && (
        <div className={classes.content}>
          <Loading justify="center" />
        </div>
      )}
      {error && (
        <div className={classes.content}>
          <Typography color="error" align="center">
            {error}
          </Typography>
        </div>
      )}
      {!error && callToAction?.message && (
        <div className={classes.content}>
          <Grid container spacing={4}>
            {callToAction.message && (
              <Grid item xs={12}>
                <Typography color="textSecondary" align="center">
                  {callToAction.message}
                </Typography>
              </Grid>
            )}
            <Grid item xs={12}>
              <Grid container justify="center" spacing={2}>
                {callToAction.buttons?.map((cta, idx) => (
                  <Grid item key={idx}>
                    <Button
                      variant="contained"
                      onClick={cta.onClick}
                      component={cta.href ? NakedLink : "button"}
                      href={cta.href}
                      as={cta.as}
                    >
                      {cta.label}
                    </Button>
                  </Grid>
                ))}
              </Grid>
            </Grid>
          </Grid>
        </div>
      )}
    </>
  );

  if (paper) {
    innerEl = (
      <Paper className={classes.paper} variant="outlined">
        {innerEl}
      </Paper>
    );
  }

  if (note) {
    innerEl = (
      <>
        {innerEl}
        {!error && !callToAction?.message && note && (
          <div className={classes.note}>
            <Typography variant="body2" color="textSecondary" align="center">
              {note}
            </Typography>
          </div>
        )}
      </>
    );
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
