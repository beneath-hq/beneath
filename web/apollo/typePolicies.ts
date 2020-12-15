import { TypePolicies } from "@apollo/client";

// Data normalization config. See https://www.apollographql.com/docs/react/caching/cache-configuration/#data-normalization
// NOTE: setting keyFields to false causes objects to be embedded in the entry of their parent object, see: https://www.apollographql.com/docs/react/caching/cache-configuration/#disabling-normalization

const typePolicies: TypePolicies = {
  Usage: { keyFields: ["entityID", "label", "time"] },
  NewUserSecret: { keyFields: ["secret", ["userSecretID"]] },
  NewServiceSecret: { keyFields: ["secret", "serviceSecretID"] },
  Organization: { keyFields: ["organizationID"] },
  OrganizationMember: { keyFields: ["organizationID", "userID"] },
  PermissionsServicesStreams: { keyFields: false },
  PermissionsUsersOrganizations: { keyFields: false },
  PermissionsUsersProjects: { keyFields: false },
  PrivateOrganization: { keyFields: ["organizationID"] },
  PrivateUser: { keyFields: ["userID"] },
  Project: { keyFields: ["projectID"] },
  ProjectMember: { keyFields: ["projectID", "userID"] },
  PublicOrganization: { keyFields: ["organizationID"] },
  Service: { keyFields: ["serviceID"] },
  Stream: { keyFields: ["streamID"] },
  StreamIndex: { keyFields: ["indexID"] },
  StreamInstance: { keyFields: ["streamInstanceID"] },
  UserSecret: { keyFields: ["userSecretID"] },
};

export default typePolicies;
