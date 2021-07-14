/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { UpdateAuthTicketInput } from "./globalTypes";

// ====================================================
// GraphQL mutation operation: UpdateAuthTicket
// ====================================================

export interface UpdateAuthTicket_updateAuthTicket {
  __typename: "AuthTicket";
  authTicketID: ControlUUID;
  updatedOn: ControlTime;
}

export interface UpdateAuthTicket {
  updateAuthTicket: UpdateAuthTicket_updateAuthTicket | null;
}

export interface UpdateAuthTicketVariables {
  input: UpdateAuthTicketInput;
}
