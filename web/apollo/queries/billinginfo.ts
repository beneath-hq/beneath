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
      billingMethodID
      paymentsDriver
      driverPayload
    }
    country
    region
    companyName
    taxNumber
  }
}
`;

export const UPDATE_BILLING_INFO = gql`
  mutation UpdateBillingInfo($organizationID: UUID!, $billingMethodID: UUID, $billingPlanID: UUID!, $country: String!, $region: String, $companyName: String, $taxNumber: String){
  updateBillingInfo(organizationID: $organizationID, billingMethodID: $billingMethodID, billingPlanID: $billingPlanID, country: $country, region: $region, companyName: $companyName, taxNumber: $taxNumber ) {
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
      readQuota
      writeQuota
    }
    billingMethod {
      billingMethodID
      paymentsDriver
      driverPayload
    }
    country
    region
    companyName
    taxNumber
  }
}
`;
