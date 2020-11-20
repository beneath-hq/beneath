import gql from "graphql-tag";

export const QUERY_BILLING_INFO = gql`
  query BillingInfo($organizationID: UUID!){
    billingInfo(organizationID: $organizationID) {
      organizationID
      billingPlan {
        billingPlanID
        default
        name
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
        UIRank
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
      nextBillingTime
      lastInvoiceTime
    }
  }
`;

export const UPDATE_BILLING_DETAILS = gql`
  mutation UpdateBillingDetails($organizationID: UUID!, $country: String, $region: String, $companyName: String, $taxNumber: String){
    updateBillingDetails(organizationID: $organizationID, country: $country, region: $region, companyName: $companyName, taxNumber: $taxNumber) {
      organizationID
      country
      region
      companyName
      taxNumber
    }
  }
`;

export const UPDATE_BILLING_METHOD = gql`
  mutation UpdateBillingMethod($organizationID: UUID!, $billingMethodID: UUID){
    updateBillingMethod(organizationID: $organizationID, billingMethodID: $billingMethodID) {
      organizationID
      billingMethod {
        billingMethodID
        paymentsDriver
        driverPayload
      }
    }
  }
`;

export const UPDATE_BILLING_PLAN = gql`
  mutation UpdateBillingPlan($organizationID: UUID!, $billingPlanID: UUID!){
    updateBillingPlan(organizationID: $organizationID, billingPlanID: $billingPlanID) {
      organizationID
      billingPlan {
        billingPlanID
        default
        name
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
        UIRank
      }
    }
  }
`;
