/* tslint:disable */
/* eslint-disable */
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: UpdateMe
// ====================================================

export interface UpdateMe_updateMe_user {
  __typename: "User";
  userID: string;
  name: string;
  bio: string | null;
}

export interface UpdateMe_updateMe {
  __typename: "Me";
  userID: string;
  user: UpdateMe_updateMe_user;
  updatedOn: ControlTime;
}

export interface UpdateMe {
  updateMe: UpdateMe_updateMe;
}

export interface UpdateMeVariables {
  name?: string | null;
  bio?: string | null;
}
