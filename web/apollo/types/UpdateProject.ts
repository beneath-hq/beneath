/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: UpdateProject
// ====================================================

export interface UpdateProject_updateProject_organization {
  __typename: "Organization";
  name: string;
}

export interface UpdateProject_updateProject {
  __typename: "Project";
  projectID: string;
  displayName: string;
  site: string | null;
  description: string | null;
  photoURL: string | null;
  updatedOn: ControlTime;
  organization: UpdateProject_updateProject_organization;
}

export interface UpdateProject {
  updateProject: UpdateProject_updateProject;
}

export interface UpdateProjectVariables {
  projectID: ControlUUID;
  displayName?: string | null;
  site?: string | null;
  description?: string | null;
  photoURL?: string | null;
}
