import { Request } from "express";

import { Key } from "./entities/Key";

export type KeyRole = "personal" | "readwrite" | "readonly";

export interface IAuthenticatedUser {
  anonymous: boolean;
  key: Key;
}

export interface IAuthenticatedRequest extends Request {
  user: IAuthenticatedUser;
  logout: () => void;
}

export interface IApolloContext {
  user: IAuthenticatedUser;
}
