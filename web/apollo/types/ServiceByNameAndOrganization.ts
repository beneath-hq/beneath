/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: ServiceByNameAndOrganization
// ====================================================

export interface ServiceByNameAndOrganization_serviceByNameAndOrganization {
  __typename: "Service";
  serviceID: ControlUUID;
  name: string;
  kind: string;
  readQuota: number | null;
  writeQuota: number | null;
}

export interface ServiceByNameAndOrganization {
  serviceByNameAndOrganization: ServiceByNameAndOrganization_serviceByNameAndOrganization;
}

export interface ServiceByNameAndOrganizationVariables {
  serviceName: string;
  organizationName: string;
}
