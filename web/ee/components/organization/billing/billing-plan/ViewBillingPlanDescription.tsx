import React, { FC } from "react";

import { BillingInfo_billingInfo_billingPlan } from "ee/apollo/types/BillingInfo";
import { Grid, makeStyles, Paper, Typography } from "@material-ui/core";
import VSpace from "components/VSpace";
import clsx from "clsx";

const useStyles = makeStyles((theme) => ({
  paper: {
    padding: theme.spacing(3),
    overflowX: "auto",
    width: "inherit",
    "&:hover": {
      backgroundColor: theme.palette.primary.dark,
      borderColor: theme.palette.primary.dark
    },
  },
  selectedPaper: {
    backgroundColor: theme.palette.primary.dark,
    borderColor: theme.palette.primary.dark
  }
}));

interface Props {
  billingPlan: BillingInfo_billingInfo_billingPlan;
  current?: boolean;
  selected?: boolean;
}

const ViewBillingPlanDescription: FC<Props> = ({billingPlan, current, selected}) => {
  const classes = useStyles();
  const currencyFormatter = new Intl.NumberFormat('en-US', {style: 'currency', currency: 'USD'});

  const isOverage = (billingPlan.readOveragePriceCents > 0) || (billingPlan.writeOveragePriceCents > 0) || (billingPlan.scanOveragePriceCents > 0);

  return (
    <>
      <Paper className={clsx(classes.paper, selected && classes.selectedPaper)} variant="outlined">
        <Grid container alignItems="center" spacing={2} justify="space-between">
          <Grid item>
            <Grid container spacing={2}>
              <Grid item>
                <Typography variant="h3">{billingPlan.name}</Typography>
              </Grid>
              {current && (
                <Grid item>
                  <Typography>
                    Current
                  </Typography>
                </Grid>
              )}
            </Grid>
          </Grid>
          <Grid item>
            <Typography>
              {(isOverage ? "starting at " : "") + currencyFormatter.format(billingPlan.basePriceCents / 100)} / month
            </Typography>
          </Grid>
        </Grid>
        <VSpace units={3} />
        <Typography>{billingPlan.description}</Typography>
      </Paper>
    </>
  );
};

export default ViewBillingPlanDescription;