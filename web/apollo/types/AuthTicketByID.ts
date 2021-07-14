/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: AuthTicketByID
// ====================================================

export interface AuthTicketByID_authTicketByID {
  __typename: "AuthTicket";
  authTicketID: ControlUUID;
  requesterName: string;
  createdOn: ControlTime;
  updatedOn: ControlTime;
}

export interface AuthTicketByID {
  authTicketByID: AuthTicketByID_authTicketByID;
}

export interface AuthTicketByIDVariables {
  authTicketID: ControlUUID;
}
