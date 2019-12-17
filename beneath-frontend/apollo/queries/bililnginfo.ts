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
    paymentsDriver
  }
}
`;