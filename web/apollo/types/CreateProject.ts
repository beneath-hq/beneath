/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { CreateProjectInput } from "./globalTypes";

// ====================================================
// GraphQL mutation operation: CreateProject
// ====================================================

export interface CreateProject_createProject_organization {
  __typename: "PublicOrganization";
  organizationID: string;
  name: string;
}

export interface CreateProject_createProject {
  __typename: "Project";
  projectID: string;
  name: string;
  displayName: string;
  public: boolean;
  description: string | null;
  site: string | null;
  photoURL: string | null;
  updatedOn: ControlTime;
  organization: CreateProject_createProject_organization;
}

export interface CreateProject {
  createProject: CreateProject_createProject;
}

export interface CreateProjectVariables {
  input: CreateProjectInput;
}
