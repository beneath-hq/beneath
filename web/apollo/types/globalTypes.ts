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
  Table = "Table",
  TableInstance = "TableInstance",
  User = "User",
}

export enum TableSchemaKind {
  Avro = "Avro",
  GraphQL = "GraphQL",
}

export enum UsageLabel {
  Hourly = "Hourly",
  Monthly = "Monthly",
  QuotaMonth = "QuotaMonth",
  Total = "Total",
}

export interface CompileSchemaInput {
  schemaKind: TableSchemaKind;
  schema: string;
  indexes?: string | null;
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

export interface CreateServiceInput {
  organizationName: string;
  projectName: string;
  serviceName: string;
  description?: string | null;
  sourceURL?: string | null;
  readQuota?: number | null;
  writeQuota?: number | null;
  scanQuota?: number | null;
  updateIfExists?: boolean | null;
}

export interface CreateTableInput {
  organizationName: string;
  projectName: string;
  tableName: string;
  schemaKind: TableSchemaKind;
  schema: string;
  indexes?: string | null;
  description?: string | null;
  meta?: boolean | null;
  allowManualWrites?: boolean | null;
  useLog?: boolean | null;
  useIndex?: boolean | null;
  useWarehouse?: boolean | null;
  logRetentionSeconds?: number | null;
  indexRetentionSeconds?: number | null;
  warehouseRetentionSeconds?: number | null;
  updateIfExists?: boolean | null;
}

export interface CreateTableInstanceInput {
  tableID: ControlUUID;
  version?: number | null;
  makePrimary?: boolean | null;
  updateIfExists?: boolean | null;
}

export interface GetEntityUsageInput {
  entityID: ControlUUID;
  label: UsageLabel;
  from?: ControlTime | null;
  until?: ControlTime | null;
}

export interface GetUsageInput {
  entityKind: EntityKind;
  entityID: ControlUUID;
  label: UsageLabel;
  from?: ControlTime | null;
  until?: ControlTime | null;
}

export interface UpdateAuthTicketInput {
  authTicketID: ControlUUID;
  approve: boolean;
}

export interface UpdateProjectInput {
  projectID: ControlUUID;
  displayName?: string | null;
  public?: boolean | null;
  description?: string | null;
  site?: string | null;
  photoURL?: string | null;
}

export interface UpdateServiceInput {
  organizationName: string;
  projectName: string;
  serviceName: string;
  description?: string | null;
  sourceURL?: string | null;
  readQuota?: number | null;
  writeQuota?: number | null;
  scanQuota?: number | null;
}

export interface UpdateTableInstanceInput {
  tableInstanceID: ControlUUID;
  makeFinal?: boolean | null;
  makePrimary?: boolean | null;
}

//==============================================================
// END Enums and Input Objects
//==============================================================
