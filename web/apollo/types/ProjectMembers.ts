/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: ProjectMembers
// ====================================================

export interface ProjectMembers_projectMembers {
  __typename: "ProjectMember";
  projectID: string;
  userID: string;
  name: string;
  displayName: string;
  photoURL: string | null;
  view: boolean;
  create: boolean;
  admin: boolean;
}

export interface ProjectMembers {
  projectMembers: ProjectMembers_projectMembers[];
}

export interface ProjectMembersVariables {
  projectID: ControlUUID;
}
