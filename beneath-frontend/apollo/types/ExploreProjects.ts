/* tslint:disable */
/* eslint-disable */
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: ExploreProjects
// ====================================================

export interface ExploreProjects_exploreProjects {
  __typename: "Project";
  projectID: string;
  name: string;
  displayName: string;
  description: string | null;
  photoURL: string | null;
  createdOn: ControlTime;
  updatedOn: ControlTime;
}

export interface ExploreProjects {
  exploreProjects: ExploreProjects_exploreProjects[];
}
