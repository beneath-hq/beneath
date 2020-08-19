/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: StreamInstancesByOrganizationProjectAndStreamName
// ====================================================

export interface StreamInstancesByOrganizationProjectAndStreamName_streamInstancesByOrganizationProjectAndStreamName {
  __typename: "StreamInstance";
  streamInstanceID: string;
  streamID: ControlUUID;
  version: number;
  createdOn: ControlTime;
  madePrimaryOn: ControlTime | null;
  madeFinalOn: ControlTime | null;
}

export interface StreamInstancesByOrganizationProjectAndStreamName {
  streamInstancesByOrganizationProjectAndStreamName: StreamInstancesByOrganizationProjectAndStreamName_streamInstancesByOrganizationProjectAndStreamName[];
}

export interface StreamInstancesByOrganizationProjectAndStreamNameVariables {
  organizationName: string;
  projectName: string;
  streamName: string;
}
