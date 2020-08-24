/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: StageProject
// ====================================================

export interface StageProject_stageProject_organization {
  __typename: "PublicOrganization";
  organizationID: string;
  name: string;
}

export interface StageProject_stageProject {
  __typename: "Project";
  projectID: string;
  name: string;
  displayName: string;
  public: boolean;
  description: string | null;
  site: string | null;
  photoURL: string | null;
  updatedOn: ControlTime;
  organization: StageProject_stageProject_organization;
}

export interface StageProject {
  stageProject: StageProject_stageProject;
}

export interface StageProjectVariables {
  organizationName: string;
  projectName: string;
  displayName?: string | null;
  public?: boolean | null;
  description?: string | null;
  site?: string | null;
  photoURL?: string | null;
}
