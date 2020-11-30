import { useQuery } from "@apollo/client";
import _ from "lodash";
import React, { FC } from "react";

import { QUERY_BILLING_PLANS } from "ee/apollo/queries/billingPlan";
import { BillingPlans } from "ee/apollo/types/BillingPlans";
import { BillingInfo_billingInfo, BillingInfo_billingInfo_billingPlan } from "ee/apollo/types/BillingInfo";
import ViewBillingPlanDescription from "./ViewBillingPlanDescription";
import ViewBillingPlanQuotas from "./ViewBillingPlanQuotas";
import { Grid, List, ListItem, makeStyles } from "@material-ui/core";
import VSpace from "components/VSpace";

const useStyles = makeStyles((theme) => ({
  list: {
    display: 'flex',
    flexDirection: 'row',
    padding: theme.spacing(0),
  },
  listItem: {
    padding: theme.spacing(0),
    marginLeft: theme.spacing(2),
    marginRight: theme.spacing(2),
    borderRadius: "4px",
  },
}));

interface Props {
  selectBillingPlan: (value: BillingInfo_billingInfo_billingPlan | null) => void;
  selectedBillingPlan: BillingInfo_billingInfo_billingPlan | null;
  billingInfo: BillingInfo_billingInfo;
}

const SelectBillingPlan: FC<Props> = ({selectBillingPlan, selectedBillingPlan, billingInfo}) => {
  const classes = useStyles();
  const { loading, error, data } = useQuery<BillingPlans>(QUERY_BILLING_PLANS, {
    context: { ee: true },
  });

  if (!data || error) return null;

  const sortedBillingPlans = data.billingPlans.slice().sort((a, b) => {
    if (a.UIRank && b.UIRank) {
      return (a.UIRank > b.UIRank ? 1 : -1);
    } else {
      return 0;
    }
  });

  return (
    <>
      <List className={classes.list}>
        {sortedBillingPlans.map((billingPlan) => (
          <React.Fragment key={billingPlan.billingPlanID}>
            <ListItem
              button
              selected={selectedBillingPlan === billingPlan}
              onClick={() => selectBillingPlan(billingPlan)}
              className={classes.listItem}
            >
              <ViewBillingPlanDescription
                billingPlan={billingPlan}
                selectable
                selected={selectedBillingPlan === billingPlan}
                current={(billingInfo.billingPlan.billingPlanID === billingPlan.billingPlanID)}
              />
            </ListItem>
          </React.Fragment>
        ))}
      </List>
      {selectedBillingPlan && (
        <>
          <VSpace units={6} />
          <Grid container>
            <Grid item xs={2}>
            </Grid>
            <Grid item xs={8}>
              <ViewBillingPlanQuotas billingPlan={selectedBillingPlan} />
            </Grid>
          </Grid>
        </>
      )}
    </>
  );
};

export default SelectBillingPlan;
