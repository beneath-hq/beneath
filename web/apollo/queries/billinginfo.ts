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
      driverPayload
    }
  }
}
`;

export const UPDATE_BILLING_INFO = gql`
  mutation UpdateBillingInfo($organizationID: UUID!, $billingMethodID: UUID! $billingPlanID: UUID!){
  updateBillingInfo(organizationID: $organizationID, billingMethodID: $billingMethodID, billingPlanID: $billingPlanID ) {
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
      driverPayload
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
      driverPayload
    }
  }
}
`;