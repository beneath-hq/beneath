import { Request } from "express";

export interface IAuthenticatedRequest extends Request {
  user: { userId: string, kind: "session"|"secret" };
  logout: () => void;
}
