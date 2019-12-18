/* tslint:disable */
/* eslint-disable */
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: UpdateMe
// ====================================================

export interface UpdateMe_updateMe_user {
  __typename: "User";
  userID: string;
  username: string;
  name: string;
  bio: string | null;
  photoURL: string | null;
}

export interface UpdateMe_updateMe {
  __typename: "Me";
  userID: string;
  readUsage: number;
  readQuota: number;
  writeUsage: number;
  writeQuota: number;
  updatedOn: ControlTime;
  user: UpdateMe_updateMe_user;
}

export interface UpdateMe {
  updateMe: UpdateMe_updateMe;
}

export interface UpdateMeVariables {
  username?: string | null;
  name?: string | null;
  bio?: string | null;
}
