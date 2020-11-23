import { useQuery } from "@apollo/client";
import _ from "lodash";
import React, { FC } from "react";

import { QUERY_BILLING_PLANS } from "ee/apollo/queries/billingPlan";
import { BillingPlans } from "ee/apollo/types/BillingPlans";
import { BillingInfo_billingInfo, BillingInfo_billingInfo_billingPlan } from "ee/apollo/types/BillingInfo";
import ViewBillingPlanDescription from "./ViewBillingPlanDescription";
import ViewBillingPlanQuotas from "./ViewBillingPlanQuotas";
import { List, ListItem, makeStyles } from "@material-ui/core";
import VSpace from "components/VSpace";

const useStyles = makeStyles((theme) => ({
  listItem: {
    padding: "0px",
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

  if (!data) {
    return <></>;
  }

  return (
    <>
      {data.billingPlans.map((billingPlan) => (
        <React.Fragment key={billingPlan.billingPlanID}>
          <List>
            <ListItem
              button
              selected={selectedBillingPlan === billingPlan}
              onClick={() => selectBillingPlan(billingPlan)}
              className={classes.listItem}
            >
              <ViewBillingPlanDescription
                billingPlan={billingPlan}
                current={(billingInfo.billingPlan.billingPlanID === billingPlan.billingPlanID)}
                selected={selectedBillingPlan === billingPlan}
              />
            </ListItem>
          </List>
        </React.Fragment>
      ))}
      {selectedBillingPlan && (
        <>
          <VSpace units={3} />
          <ViewBillingPlanQuotas billingPlan={selectedBillingPlan} />
        </>
      )}
    </>
  );
};

export default SelectBillingPlan;
