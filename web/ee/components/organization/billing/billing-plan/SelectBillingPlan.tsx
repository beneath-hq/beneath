import { useQuery } from "@apollo/client";
import _ from "lodash";
import { Link, List, ListItem, makeStyles } from "@material-ui/core";
import { Alert } from "@material-ui/lab";
import React, { FC } from "react";

import { QUERY_BILLING_PLANS } from "ee/apollo/queries/billingPlan";
import { BillingInfo_billingInfo, BillingInfo_billingInfo_billingPlan } from "ee/apollo/types/BillingInfo";
import { BillingPlans } from "ee/apollo/types/BillingPlans";
import { CONTACT_LINK } from "ee/lib/billing";
import VSpace from "components/VSpace";
import ViewBillingPlanDescription from "./ViewBillingPlanDescription";
import ViewBillingPlanQuotas from "./ViewBillingPlanQuotas";
import clsx from "clsx";

const useStyles = makeStyles((theme) => ({
  list: {
    display: "flex",
    padding: theme.spacing(0),
    [theme.breakpoints.down("xs")]: {
      flexDirection: "column",
    },
    [theme.breakpoints.up("sm")]: {
      flexDirection: "row",
    },
  },
  listItem: {
    padding: theme.spacing(0),
    marginLeft: theme.spacing(2),
    [theme.breakpoints.down("xs")]: {
      marginLeft: theme.spacing(0),
      marginTop: theme.spacing(2),
    },
    borderRadius: "4px",
  },
  firstListItem: {
    [theme.breakpoints.down("xs")]: {
      marginTop: theme.spacing(0),
    },
    [theme.breakpoints.up("sm")]: {
      marginLeft: theme.spacing(0),
    },
  },
}));

interface Props {
  selectBillingPlan: (value: BillingInfo_billingInfo_billingPlan | null) => void;
  selectedBillingPlan: BillingInfo_billingInfo_billingPlan | null;
  billingInfo: BillingInfo_billingInfo;
}

const SelectBillingPlan: FC<Props> = ({ selectBillingPlan, selectedBillingPlan, billingInfo }) => {
  const classes = useStyles();
  const { loading, error, data } = useQuery<BillingPlans>(QUERY_BILLING_PLANS, {
    context: { ee: true },
  });

  if (!data || error) return null;

  const sortedBillingPlans = data.billingPlans.slice().sort((a, b) => {
    if (a.UIRank && b.UIRank) {
      return a.UIRank > b.UIRank ? 1 : -1;
    } else {
      return 0;
    }
  });

  return (
    <>
      <Alert severity="info">
        <Link href={CONTACT_LINK} target="_blank">
          Contact sales
        </Link>{" "}
        for Enterprise plans, which offer custom pricing, team billing, premium support, and more
      </Alert>
      <VSpace units={3} />
      <List className={classes.list}>
        {sortedBillingPlans.map((billingPlan, idx) => (
          <React.Fragment key={billingPlan.billingPlanID}>
            <ListItem
              button
              selected={selectedBillingPlan?.billingPlanID === billingPlan.billingPlanID}
              onClick={() => selectBillingPlan(billingPlan)}
              className={clsx(classes.listItem, idx === 0 && classes.firstListItem)}
            >
              <ViewBillingPlanDescription
                billingPlan={billingPlan}
                selectable
                selected={selectedBillingPlan?.billingPlanID === billingPlan.billingPlanID}
                current={billingInfo.billingPlan.billingPlanID === billingPlan.billingPlanID}
              />
            </ListItem>
          </React.Fragment>
        ))}
      </List>
      {selectedBillingPlan && (
        <>
          <VSpace units={4} />
          <ViewBillingPlanQuotas billingPlan={selectedBillingPlan} />
        </>
      )}
    </>
  );
};

export default SelectBillingPlan;
