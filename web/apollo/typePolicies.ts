import { TypePolicies } from "@apollo/client";

// Data normalization config. See https://www.apollographql.com/docs/react/caching/cache-configuration/#data-normalization
// NOTE: setting keyFields to false causes objects to be embedded in the entry of their parent object, see: https://www.apollographql.com/docs/react/caching/cache-configuration/#disabling-normalization

const typePolicies: TypePolicies = {
  Usage: { keyFields: ["entityID", "label", "time"] },
  NewUserSecret: { keyFields: ["secret", ["userSecretID"]] },
  NewServiceSecret: { keyFields: ["secret", ["serviceSecretID"]] },
  Organization: { keyFields: ["organizationID"] },
  OrganizationMember: { keyFields: ["organizationID", "userID"] },
  PermissionsServicesTables: { keyFields: ["serviceID", "tableID"] },
  PermissionsUsersOrganizations: { keyFields: false },
  PermissionsUsersProjects: { keyFields: false },
  PrivateOrganization: { keyFields: ["organizationID"] },
  PrivateUser: { keyFields: ["userID"] },
  Project: { keyFields: ["projectID"] },
  ProjectMember: { keyFields: ["projectID", "userID"] },
  PublicOrganization: { keyFields: ["organizationID"] },
  Service: { keyFields: ["serviceID"] },
  Table: { keyFields: ["tableID"] },
  TableIndex: { keyFields: ["indexID"] },
  TableInstance: { keyFields: ["tableInstanceID"] },
  UserSecret: { keyFields: ["userSecretID"] },
};

export default typePolicies;
