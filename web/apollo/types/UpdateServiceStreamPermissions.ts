/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: UpdateServiceStreamPermissions
// ====================================================

export interface UpdateServiceStreamPermissions_updateServiceStreamPermissions_stream_project_organization {
  __typename: "PublicOrganization";
  organizationID: string;
  name: string;
}

export interface UpdateServiceStreamPermissions_updateServiceStreamPermissions_stream_project {
  __typename: "Project";
  projectID: string;
  name: string;
  organization: UpdateServiceStreamPermissions_updateServiceStreamPermissions_stream_project_organization;
}

export interface UpdateServiceStreamPermissions_updateServiceStreamPermissions_stream {
  __typename: "Stream";
  streamID: string;
  name: string;
  project: UpdateServiceStreamPermissions_updateServiceStreamPermissions_stream_project;
}

export interface UpdateServiceStreamPermissions_updateServiceStreamPermissions {
  __typename: "PermissionsServicesStreams";
  serviceID: ControlUUID;
  streamID: ControlUUID;
  read: boolean;
  write: boolean;
  stream: UpdateServiceStreamPermissions_updateServiceStreamPermissions_stream | null;
}

export interface UpdateServiceStreamPermissions {
  updateServiceStreamPermissions: UpdateServiceStreamPermissions_updateServiceStreamPermissions;
}

export interface UpdateServiceStreamPermissionsVariables {
  serviceID: ControlUUID;
  streamID: ControlUUID;
  read?: boolean | null;
  write?: boolean | null;
}
