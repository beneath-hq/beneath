import React, { FC } from "react";

import { BillingInfo_billingInfo_billingPlan } from "ee/apollo/types/BillingInfo";
import { Grid, makeStyles, Paper, Typography } from "@material-ui/core";
import VSpace from "components/VSpace";
import clsx from "clsx";

const useStyles = makeStyles((theme) => ({
  paper: {
    padding: theme.spacing(3),
    overflowX: "auto",
    height: "100%",
  },
  selectablePaper: {
    "&:hover": {
      backgroundColor: theme.palette.primary.dark,
      borderColor: theme.palette.primary.dark
    },
  },
  selectedPaper: {
    backgroundColor: theme.palette.primary.dark,
    borderColor: theme.palette.primary.dark
  },
}));

interface Props {
  billingPlan: BillingInfo_billingInfo_billingPlan;
  selectable?: boolean;
  current?: boolean;
  selected?: boolean;
}

const ViewBillingPlanDescription: FC<Props> = ({billingPlan, selectable, selected, current}) => {
  const classes = useStyles();
  const currencyFormatter = new Intl.NumberFormat('en-US', {style: 'currency', currency: 'USD'});

  const isOverage = (billingPlan.readOveragePriceCents > 0) || (billingPlan.writeOveragePriceCents > 0) || (billingPlan.scanOveragePriceCents > 0);

  return (
    <>
      <Paper className={clsx(classes.paper, selectable && classes.selectablePaper, selected && classes.selectedPaper)} variant="outlined">
        {!selectable && (
          <>
            <Typography variant="h2" gutterBottom>
              Billing Plan
            </Typography>
            <Typography variant="body2" color="textSecondary">
              Your new billing plan to begin immediately upon purchase
            </Typography>
            <VSpace units={3} />
          </>
        )}
        <Grid container spacing={2} alignItems="center">
          <Grid item>
            <Typography><strong>{billingPlan.name}</strong></Typography>
          </Grid>
          {current && (
            <Grid item >
              <Typography>
                (Current plan)
              </Typography>
            </Grid>
          )}
        </Grid>
        <VSpace units={1} />
        <Typography>
          {(isOverage ? "starting at " : "") + currencyFormatter.format(billingPlan.basePriceCents / 100)} / month
        </Typography>
        <VSpace units={3} />
        <Typography>{billingPlan.description}</Typography>
      </Paper>
    </>
  );
};

export default ViewBillingPlanDescription;