import { Button, Grid, makeStyles, Step, StepContent, StepLabel, Stepper, Theme, Typography } from "@material-ui/core";
import React, { FC } from "react";

import ContentContainer from "components/ContentContainer";
import { NakedLink } from "components/Link";
import { setRedirectAfterAuth } from "lib/authRedirect";
import ExploreProjectsTiles from "./ExploreProjectsTiles";
import Tile from "./tiles/Tile";

const useStyles = makeStyles((theme: Theme) => ({
  sectionTitle: {
    marginTop: theme.spacing(4),
  },
  title: {
    fontSize: theme.typography.pxToRem(32),
  },
  stepperCallToActionButton: {
    marginLeft: theme.spacing(3),
    marginBottom: theme.spacing(3),
  },
}));

const Welcome: FC = () => {
  const classes = useStyles();

  const steps = [
    {
      label: "Create an account",
      content: "Connect with Github or Google, and start writing data for free – no credit card required",
    },
    { label: "Create a stream", content: `Provide a name and schema for your new stream` },
    {
      label: "Write data",
      content: "Use the Python client or the web console to write records to your stream",
    },
  ];

  return (
    <Grid container spacing={3}>
      <Grid item xs={12}>
        <Typography variant="h1" className={classes.title}>
          Welcome to Beneath!
        </Typography>
      </Grid>
      <Grid item xs={12}>
        <Typography variant="h3" className={classes.sectionTitle}>
          Explore the public streams...
        </Typography>
      </Grid>
      <ExploreProjectsTiles />
      <Grid item xs={12}>
        <Typography variant="h3" className={classes.sectionTitle}>
          ... or create your first stream
        </Typography>
      </Grid>
      <Tile shape="full">
        <Grid container direction="column">
          <Grid item>
            <Stepper orientation="vertical">
              {steps.map(({ label, content }, index) => (
                <Step key={index} active={true}>
                  <StepLabel>
                    <Typography variant="h4">{label}</Typography>
                  </StepLabel>
                  <StepContent>
                    <Typography variant="body2">{content}</Typography>
                  </StepContent>
                </Step>
              ))}
            </Stepper>
          </Grid>
          <Grid item>
            <Button
              variant="contained"
              color="primary"
              component={NakedLink}
              onClick={() => setRedirectAfterAuth("/-/create/stream")}
              href="/-/auth"
              className={classes.stepperCallToActionButton}
            >
              Get started!
            </Button>
          </Grid>
        </Grid>
      </Tile>
    </Grid>
  );
};

export default Welcome;
