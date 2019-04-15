import { Request } from "express";

export type AuthenticatedUserScope = "modify";

export interface IAuthenticatedUser {
  userId: string;
  kind: "anonymous" | "secret" | "session";
  scopes: AuthenticatedUserScope[];
}

export interface IAuthenticatedRequest extends Request {
  user: IAuthenticatedUser;
  logout: () => void;
}

export interface IApolloContext {
  user: IAuthenticatedUser;
}
