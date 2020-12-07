/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

//==============================================================
// START Enums and Input Objects
//==============================================================

export enum EntityKind {
  Organization = "Organization",
  Service = "Service",
  Stream = "Stream",
  StreamInstance = "StreamInstance",
  User = "User",
}

export enum StreamSchemaKind {
  GraphQL = "GraphQL",
}

export interface CreateProjectInput {
  organizationID: ControlUUID;
  projectName: string;
  displayName?: string | null;
  public?: boolean | null;
  description?: string | null;
  site?: string | null;
  photoURL?: string | null;
}

export interface UpdateProjectInput {
  projectID: ControlUUID;
  displayName?: string | null;
  public?: boolean | null;
  description?: string | null;
  site?: string | null;
  photoURL?: string | null;
}

//==============================================================
// END Enums and Input Objects
//==============================================================
