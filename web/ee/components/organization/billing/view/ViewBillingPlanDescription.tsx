import React, { FC } from "react";

import { BillingInfo_billingInfo_billingPlan } from "ee/apollo/types/BillingInfo";
import { Grid, makeStyles, Paper, Typography } from "@material-ui/core";
import { prettyPrintBytes } from "components/metrics/util";
import VSpace from "components/VSpace";

const useStyles = makeStyles((theme) => ({
  paperPadding: {
    padding: theme.spacing(3)
  }
}));

interface Props {
  billingPlan: BillingInfo_billingInfo_billingPlan;
}

const ViewBillingPlanDescription: FC<Props> = ({billingPlan}) => {
  const classes = useStyles();

  const rows = [
    {key: "Plan name", value: billingPlan.description},
    {key: "Price", value: `$${billingPlan.basePriceCents / 100}`},
    {key: "Read quota", value: prettyPrintBytes(billingPlan.readQuota)},
    {key: "Write quota", value: prettyPrintBytes(billingPlan.writeQuota)},
    {key: "Scan quota", value: prettyPrintBytes(billingPlan.scanQuota)},
  ];

  return (
    <>
      <Paper className={classes.paperPadding} variant="outlined">
        {rows.map((row) => (
          <React.Fragment key={row.key}>
            <Grid container justify="space-between" alignItems="center">
              <Grid item>
                <Typography>
                  {row.key}:
                </Typography>
              </Grid>
              <Grid item>
                <Typography>
                  {row.value}
                </Typography>
              </Grid>
            </Grid>
            <VSpace units={1} />
          </React.Fragment>
        ))}
      </Paper>
    </>
  );
};

export default ViewBillingPlanDescription;

// export const PROFESSIONAL = "feature1";
// export const PROFESSIONAL_BOOST = "feature1";

// {[
//               "5 GB writes included in base. Then $2/GB.",
//               "25 GB reads included in base. Then $1/GB.",
//               "Private projects",
//               "Role-based access controls",
//             ].map((feature) => {
//               return (
//                 <React.Fragment key={feature}>
//                   <ListItem dense>
//                     <CheckIcon className={classes.icon} />
//                     <Typography component="span">{feature}</Typography>
//                   </ListItem>
//                 </React.Fragment>
//               );
//             })}