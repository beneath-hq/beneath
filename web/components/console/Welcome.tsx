import { Button, Grid, makeStyles, Paper, Step, StepButton, StepContent, StepLabel, Stepper, Theme, Typography } from "@material-ui/core";
import ContentContainer from "components/ContentContainer";
import { NakedLink } from "components/Link";
import React, { FC } from "react";

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
    marginBottom: theme.spacing(3)
  }
}));

const Welcome: FC = () => {
  const classes = useStyles();

  const steps = [
    {label: 'Create an account', content: 'Connect with Github or Google, and you can start writing data for free - no credit card required'},
    {label: 'Create a stream', content: `Provide the name, description, and schema for your new stream`},
    {label: 'Write data', content: 'Use the easy Beneath Python library or the Beneath UI to write records to your stream'}
  ];

  return (
    <ContentContainer maxWidth="md" >
      <Grid container spacing={3}>
        <Grid item xs={12}>
          <Typography variant="h1" className={classes.title}>
            Welcome to Beneath!
          </Typography>
        </Grid>

        <Grid item xs={12}>
          <Typography variant="h3" className={classes.sectionTitle}>
            Three quick steps to create your first stream
          </Typography>
        </Grid>
        <Tile shape="full">
          <Grid container direction="column">
            <Grid item>
              <Stepper orientation="vertical">
                {steps.map(({label, content}, index) => (
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
              {/* TODO: after auth, redirect to the Create Stream page */}
              <Button variant="contained" color="primary" component={NakedLink} href="/-/auth" className={classes.stepperCallToActionButton}
              >
                Get started!
              </Button>
            </Grid>
          </Grid>
        </Tile>

        <Grid item xs={12}>
          <Typography variant="h3" className={classes.sectionTitle}>
            Featured projects and tutorials
          </Typography>
        </Grid>
        <ExploreProjectsTiles />
      </Grid>
    </ContentContainer>
  );
};

export default Welcome;
