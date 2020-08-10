/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: ServiceByOrganizationProjectAndName
// ====================================================

export interface ServiceByOrganizationProjectAndName_serviceByOrganizationProjectAndName {
  __typename: "Service";
  serviceID: ControlUUID;
  name: string;
  description: string | null;
  sourceURL: string | null;
  readQuota: number | null;
  writeQuota: number | null;
  scanQuota: number | null;
}

export interface ServiceByOrganizationProjectAndName {
  serviceByOrganizationProjectAndName: ServiceByOrganizationProjectAndName_serviceByOrganizationProjectAndName;
}

export interface ServiceByOrganizationProjectAndNameVariables {
  organizationName: string;
  projectName: string;
  serviceName: string;
}
