import { AuthenticationError, ForbiddenError } from "apollo-server";

import { Key, KeyRole } from "../entities/Key";
import { Project } from "../entities/Project";
import { User } from "../entities/User";

export class Auth {
  public key: Key;

  constructor(key: Key) {
    this.key = key;
  }

  public getUserId(): string {
    if (this.key && this.key.userId) {
      return this.key.userId;
    }
    return undefined;
  }

  public getShallowUser(): User {
    if (this.key && this.key.user) {
      return this.key.user;
    } else if (this.key && this.key.userId) {
      const user = new User();
      user.userId = this.key.userId;
      return user;
    }
    return undefined;
  }

  public isAnonymous(): boolean {
    return !this.key;
  }

  public requireAnonymous() {
    if (!this.isAnonymous()) {
      throw new AuthenticationError("Must not be authenticated");
    }
  }

  public isNotAnonymous(): boolean {
    return !!this.key;
  }

  public requireNotAnonymous() {
    if (!this.isNotAnonymous()) {
      throw new AuthenticationError("Must be authenticated");
    }
  }

  public isPersonalUser(): boolean  {
    return this.key && this.key.userId && this.key.role === KeyRole.Manage;
  }

  public requirePersonalUser() {
    this.requireNotAnonymous();
    if (!this.isPersonalUser()) {
      throw new ForbiddenError("Only permitted with personal login");
    }
  }

  public canEditUser(userId: string): boolean {
    return this.key && this.key.userId === userId;
  }

  public requireCanEditUser(userId: string) {
    this.requirePersonalUser();
    if (!this.canEditUser(userId)) {
      throw new ForbiddenError("Can only edit yourself");
    }
  }

  public async canReadProject(projectId: string): Promise<boolean> {
    // TODO
    return true;
  }

  public async requireCanReadProject(projectId: string) {
    // TODO
    return await this.canReadProject(projectId);
  }

  public async canEditProject(projectId: string): Promise<boolean> {
    return this.isPersonalUser() && await Project.isUserInProject(this.key.userId, projectId);
  }

  public async requireCanEditProject(projectId: string) {
    this.requirePersonalUser();
    if (!(await this.canEditProject(projectId))) {
      throw new ForbiddenError("Must be member of project to edit");
    }
  }

}
