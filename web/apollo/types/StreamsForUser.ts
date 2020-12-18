/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: StreamsForUser
// ====================================================

export interface StreamsForUser_streamsForUser_project_organization {
  __typename: "PublicOrganization";
  organizationID: string;
  name: string;
}

export interface StreamsForUser_streamsForUser_project {
  __typename: "Project";
  projectID: string;
  name: string;
  public: boolean;
  organization: StreamsForUser_streamsForUser_project_organization;
}

export interface StreamsForUser_streamsForUser {
  __typename: "Stream";
  streamID: string;
  name: string;
  description: string | null;
  createdOn: ControlTime;
  updatedOn: ControlTime;
  project: StreamsForUser_streamsForUser_project;
}

export interface StreamsForUser {
  streamsForUser: StreamsForUser_streamsForUser[];
}

export interface StreamsForUserVariables {
  userID: ControlUUID;
}
