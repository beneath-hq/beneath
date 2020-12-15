/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { UpdateServiceInput } from "./globalTypes";

// ====================================================
// GraphQL mutation operation: UpdateService
// ====================================================

export interface UpdateService_updateService_project_organization {
  __typename: "PublicOrganization";
  organizationID: string;
  name: string;
}

export interface UpdateService_updateService_project {
  __typename: "Project";
  projectID: string;
  name: string;
  organization: UpdateService_updateService_project_organization;
}

export interface UpdateService_updateService {
  __typename: "Service";
  serviceID: ControlUUID;
  name: string;
  description: string | null;
  sourceURL: string | null;
  readQuota: number | null;
  writeQuota: number | null;
  scanQuota: number | null;
  project: UpdateService_updateService_project;
}

export interface UpdateService {
  updateService: UpdateService_updateService;
}

export interface UpdateServiceVariables {
  input: UpdateServiceInput;
}
