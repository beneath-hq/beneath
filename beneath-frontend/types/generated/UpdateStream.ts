/* tslint:disable */
/* eslint-disable */
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: UpdateStream
// ====================================================

export interface UpdateStream_updateStream {
  __typename: "Stream";
  streamID: string;
  description: string | null;
  manual: boolean;
}

export interface UpdateStream {
  updateStream: UpdateStream_updateStream;
}

export interface UpdateStreamVariables {
  streamID: ControlUUID;
  description?: string | null;
  manual?: boolean | null;
}
