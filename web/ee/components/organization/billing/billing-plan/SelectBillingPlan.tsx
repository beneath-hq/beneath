import { useQuery } from "@apollo/client";
import _ from "lodash";
import { makeStyles } from "@material-ui/core/styles";
import React, { FC } from "react";

import { QUERY_BILLING_PLANS } from "ee/apollo/queries/billingPlan";
import { BillingPlans } from "ee/apollo/types/BillingPlans";
import { BillingInfo_billingInfo, BillingInfo_billingInfo_billingPlan } from "ee/apollo/types/BillingInfo";
import RadioGroup from "components/forms/RadioGroup";
import { PROFESSIONAL_PLAN, PROFESSIONAL_BOOST_PLAN } from "ee/lib/billing";
import ViewBillingPlanDescription from "./ViewBillingPlanDescription";
import VSpace from "components/VSpace";

const useStyles = makeStyles((theme) => ({
}));

interface Props {
  selectBillingPlan: (value: BillingInfo_billingInfo_billingPlan | null) => void;
  selectedBillingPlan: BillingInfo_billingInfo_billingPlan | null;
  billingInfo: BillingInfo_billingInfo;
}

const SelectBillingPlan: FC<Props> = ({selectBillingPlan, selectedBillingPlan, billingInfo}) => {
  const [selectedBillingPlanLabel, setSelectedBillingPlanLabel] = React.useState("");
  const { loading, error, data } = useQuery<BillingPlans>(QUERY_BILLING_PLANS, {
    context: { ee: true },
  });

  if (!data) {
    return <></>;
  }

  const professionalPlan = data.billingPlans.find((billingPlan) => billingPlan.description === PROFESSIONAL_PLAN);
  const professionalBoostPlan = data.billingPlans.find((billingPlan) => billingPlan.description === PROFESSIONAL_BOOST_PLAN);
  const professionalPlanLabel = "Professional".concat(billingInfo.billingPlan.billingPlanID === professionalPlan?.billingPlanID ? " (current plan)" : "");
  const professionalBoostPlanLabel = "Professional Boost".concat(billingInfo.billingPlan.billingPlanID === professionalBoostPlan?.billingPlanID ? " (current plan)" : "");

  return (
    <>
      {/* <Typography>
        Current plan: {billingInfo.billingPlan.description}
      </Typography> */}
      <RadioGroup
        options={[
          { value: PROFESSIONAL_PLAN, label: professionalPlanLabel },
          { value: PROFESSIONAL_BOOST_PLAN, label: professionalBoostPlanLabel },
        ]}
        row
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
      <VSpace units={2} />
      {selectedBillingPlan && (
        <ViewBillingPlanDescription billingPlan={selectedBillingPlan} />
      )}
      <VSpace units={2} />
      {/* <ViewBillingPlanDescription billingPlan={professionalBoostPlan as BillingInfo_billingInfo_billingPlan} /> */}
    </>
  );
};

export default SelectBillingPlan;
