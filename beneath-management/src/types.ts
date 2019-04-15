import { Request } from "express";

export interface IAuthenticatedUser {
  userId: string;
  kind: "anonymous" | "secret" | "session";
  scopes: string[];
}

export interface IAuthenticatedRequest extends Request {
  user: IAuthenticatedUser;
  logout: () => void;
}

export interface IApolloContext {
  user: IAuthenticatedUser;
}
