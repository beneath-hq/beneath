/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: StreamByOrganizationProjectAndName
// ====================================================

export interface StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName_project_organization {
  __typename: "PublicOrganization";
  name: string;
}

export interface StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName_project {
  __typename: "Project";
  projectID: string;
  name: string;
  organization: StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName_project_organization;
}

export interface StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName_streamIndexes {
  __typename: "StreamIndex";
  indexID: string;
  fields: string[];
  primary: boolean;
  normalize: boolean;
}

export interface StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName {
  __typename: "Stream";
  streamID: string;
  name: string;
  description: string | null;
  createdOn: ControlTime;
  updatedOn: ControlTime;
  project: StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName_project;
  schema: string;
  avroSchema: string;
  streamIndexes: StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName_streamIndexes[];
  external: boolean;
  batch: boolean;
  manual: boolean;
  retentionSeconds: number;
  instancesCreatedCount: number;
  instancesCommittedCount: number;
  currentStreamInstanceID: ControlUUID | null;
}

export interface StreamByOrganizationProjectAndName {
  streamByOrganizationProjectAndName: StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName;
}

export interface StreamByOrganizationProjectAndNameVariables {
  organizationName: string;
  projectName: string;
  streamName: string;
}
