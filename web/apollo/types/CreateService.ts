/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { CreateServiceInput } from "./globalTypes";

// ====================================================
// GraphQL mutation operation: CreateService
// ====================================================

export interface CreateService_createService_project_organization {
  __typename: "PublicOrganization";
  organizationID: string;
  name: string;
}

export interface CreateService_createService_project {
  __typename: "Project";
  projectID: string;
  name: string;
  organization: CreateService_createService_project_organization;
}

export interface CreateService_createService {
  __typename: "Service";
  serviceID: string;
  name: string;
  description: string | null;
  sourceURL: string | null;
  readQuota: number | null;
  writeQuota: number | null;
  scanQuota: number | null;
  project: CreateService_createService_project;
}

export interface CreateService {
  createService: CreateService_createService;
}

export interface CreateServiceVariables {
  input: CreateServiceInput;
}
