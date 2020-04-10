import gql from "graphql-tag";

export const QUERY_BILLING_INFO = gql`
  query BillingInfo($organizationID: UUID!){
  billingInfo(organizationID: $organizationID) {
    organizationID
    billingPlan {
      billingPlanID
	    description
	    currency
      period
	    seatPriceCents
	    seatReadQuota
	    seatWriteQuota
	    readOveragePriceCents
	    writeOveragePriceCents
	    baseReadQuota
	    baseWriteQuota
    }
    billingMethod {
      paymentsDriver
    }
  }
}
`;

export const UPDATE_BILLING_INFO_BILLING_METHOD = gql`
  mutation UpdateBillingInfoBillingMethod($organizationID: UUID!, $billingMethodID: UUID!){
  updateBillingInfoBillingMethod(organizationID: $organizationID, billingMethodID: $billingMethodID) {
    organizationID
    billingPlan {
      billingPlanID
	    description
	    currency
      period
	    seatPriceCents
	    seatReadQuota
	    seatWriteQuota
	    readOveragePriceCents
	    writeOveragePriceCents
	    baseReadQuota
	    baseWriteQuota
    }
    billingMethod {
      paymentsDriver
    }
  }
}
`;

export const UPDATE_BILLING_INFO_BILLING_PLAN = gql`
  mutation UpdateBillingInfoBillingPlan($organizationID: UUID!, $billingPlanID: UUID!){
  updateBillingInfoBillingPlan(organizationID: $organizationID, billingPlanID: $billingPlanID ) {
    organizationID
    billingPlan {
      billingPlanID
	    description
	    currency
      period
	    seatPriceCents
	    seatReadQuota
	    seatWriteQuota
	    readOveragePriceCents
	    writeOveragePriceCents
	    baseReadQuota
	    baseWriteQuota
    }
    billingMethod {
      paymentsDriver
    }
  }
}
`;