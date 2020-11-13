import { useQuery } from "@apollo/client";
import _ from "lodash";
import { makeStyles } from "@material-ui/core/styles";
import React, { FC } from "react";

import { QUERY_BILLING_PLANS } from "ee/apollo/queries/billingPlan";
import { BillingPlans } from "ee/apollo/types/BillingPlans";
import { Button, Grid, Typography } from "@material-ui/core";
import { BillingInfo_billingInfo, BillingInfo_billingInfo_billingPlan } from "ee/apollo/types/BillingInfo";
import RadioGroup from "components/forms/RadioGroup";
import { PROFESSIONAL_PLAN, PROFESSIONAL_BOOST_PLAN } from "ee/lib/billing";

const useStyles = makeStyles((theme) => ({
}));

interface Props {
  selectBillingPlan: (value: BillingInfo_billingInfo_billingPlan | null) => void;
  billingInfo: BillingInfo_billingInfo;
}

const SelectBillingPlan: FC<Props> = ({selectBillingPlan, billingInfo}) => {
  const [selectedBillingPlanLabel, setSelectedBillingPlanLabel] = React.useState("");
  // TODO: need to get the BillingPlanID, then I can use that in UpdateBillingInfo() in the ViewBilling/ChangeBillingPlan/Checkout component
  const { loading, error, data } = useQuery<BillingPlans>(QUERY_BILLING_PLANS, {
    context: { ee: true },
  });

  if (!data) {
    return <></>;
  }

  const professionalPlan = data.billingPlans.find((billingPlan) => billingPlan.description === PROFESSIONAL_PLAN);
  const professionalBoostPlan = data.billingPlans.find((billingPlan) => billingPlan.description === PROFESSIONAL_BOOST_PLAN);

  return (
    <>
      {/* <Typography>
        Current plan: {billingInfo.billingPlan.description}
      </Typography> */}
      <RadioGroup
        options={[
          // {value: "Free", label: "Free plan"},
          {value: PROFESSIONAL_PLAN, label: "Professional plan"},
          {value: PROFESSIONAL_BOOST_PLAN, label: "Professional Boost plan"},
          // {value: "Enterprise", label: "Enterprise plan"}
        ]}
        value={selectedBillingPlanLabel}
        onChange={(_, value) => {
          if (value === PROFESSIONAL_PLAN) {
            setSelectedBillingPlanLabel(PROFESSIONAL_PLAN);
            selectBillingPlan(professionalPlan as BillingInfo_billingInfo_billingPlan);
          } else if (value === PROFESSIONAL_BOOST_PLAN) {
            setSelectedBillingPlanLabel(PROFESSIONAL_BOOST_PLAN);
            selectBillingPlan(professionalBoostPlan as BillingInfo_billingInfo_billingPlan);
          }
        }}
      />
    </>
  );
};

export default SelectBillingPlan;
