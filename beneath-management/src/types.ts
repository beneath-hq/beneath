import { Request } from "express";

import { Key } from "./entities/Key";

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
