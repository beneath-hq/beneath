/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { UpdateProjectInput } from "./globalTypes";

// ====================================================
// GraphQL mutation operation: UpdateProject
// ====================================================

export interface UpdateProject_updateProject_organization {
  __typename: "PublicOrganization";
  organizationID: string;
  name: string;
}

export interface UpdateProject_updateProject {
  __typename: "Project";
  projectID: string;
  name: string;
  displayName: string;
  public: boolean;
  description: string | null;
  site: string | null;
  photoURL: string | null;
  updatedOn: ControlTime;
  organization: UpdateProject_updateProject_organization;
}

export interface UpdateProject {
  updateProject: UpdateProject_updateProject;
}

export interface UpdateProjectVariables {
  input: UpdateProjectInput;
}
