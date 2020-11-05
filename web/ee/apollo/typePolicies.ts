import { TypePolicies } from "@apollo/client";

// Data normalization config. See https://www.apollographql.com/docs/react/caching/cache-configuration/#data-normalization
// NOTE: setting keyFields to false causes objects to be embedded in the entry of their parent object, see: https://www.apollographql.com/docs/react/caching/cache-configuration/#disabling-normalization

const typePolicies: TypePolicies = {
  BillingInfo: { keyFields: ["organizationID"] },
  BillingMethod: { keyFields: ["billingMethodID"] },
  BillingPlan: { keyFields: ["billingPlanID"] },
};

export default typePolicies;
