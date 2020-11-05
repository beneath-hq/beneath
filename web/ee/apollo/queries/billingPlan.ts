import gql from "graphql-tag";

export const QUERY_BILLING_PLANS = gql`
  query BillingPlans{
    billingPlans {
      billingPlanID
      default
      description
      currency
      period
      basePriceCents
      seatPriceCents
      baseReadQuota
      baseWriteQuota
      baseScanQuota
      seatReadQuota
      seatWriteQuota
      seatScanQuota
      readQuota
      writeQuota
      scanQuota
      readOveragePriceCents
      writeOveragePriceCents
      scanOveragePriceCents
      availableInUI
    }
  }
`;
