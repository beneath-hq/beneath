/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: StreamPermissionsForService
// ====================================================

export interface StreamPermissionsForService_streamPermissionsForService_stream_project_organization {
  __typename: "PublicOrganization";
  organizationID: string;
  name: string;
}

export interface StreamPermissionsForService_streamPermissionsForService_stream_project {
  __typename: "Project";
  projectID: string;
  name: string;
  organization: StreamPermissionsForService_streamPermissionsForService_stream_project_organization;
}

export interface StreamPermissionsForService_streamPermissionsForService_stream {
  __typename: "Stream";
  streamID: string;
  name: string;
  project: StreamPermissionsForService_streamPermissionsForService_stream_project;
}

export interface StreamPermissionsForService_streamPermissionsForService {
  __typename: "PermissionsServicesStreams";
  serviceID: ControlUUID;
  streamID: ControlUUID;
  read: boolean;
  write: boolean;
  stream: StreamPermissionsForService_streamPermissionsForService_stream | null;
}

export interface StreamPermissionsForService {
  streamPermissionsForService: StreamPermissionsForService_streamPermissionsForService[];
}

export interface StreamPermissionsForServiceVariables {
  serviceID: ControlUUID;
}
