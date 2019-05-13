import { Request } from "express";
import { Auth } from "./lib/auth";

export interface IAuthenticatedRequest extends Request {
  auth: Auth;
}

export interface IApolloContext {
  auth: Auth;
}
