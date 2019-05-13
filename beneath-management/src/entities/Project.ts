import { IsFQDN, IsLowercase, Length, Matches } from "class-validator";
import {
  BaseEntity, Column, CreateDateColumn, Entity, getConnection, JoinTable,
  ManyToMany, OneToMany, PrimaryGeneratedColumn, UpdateDateColumn,
} from "typeorm";

import { Key } from "./Key";
import { User } from "./User";

@Entity("projects")
export class Project extends BaseEntity {

  @PrimaryGeneratedColumn("uuid", { name: "project_id" })
  public projectId: string;

  @Column({ length: 16, unique: true })
  @IsLowercase()
  @Length(3, 16)
  @Matches(/[_a-zA-Z][_\-a-zA-Z0-9]*/)
  public name: string;

  @Column({ length: 16, unique: true, name: "display_name" })
  public displayName: string;

  @Column({ length: 255 })
  @IsFQDN()
  public site: string;

  @Column({ length: 256 })
  public description: string;

  @CreateDateColumn({ name: "created_on" })
  public createdOn: Date;

  @UpdateDateColumn({ name: "updated_on" })
  public updatedOn: Date;

  @OneToMany((type) => Key, (key) => key.project)
  public keys: Key[];

  @ManyToMany((type) => User, (user) => user.projects)
  @JoinTable({
    name: "projects_users",
    joinColumn: {
      name: "project_id",
      referencedColumnName: "projectId", // note: column in entity, not in table (which is project_id)
    },
    inverseJoinColumn: {
      name: "user_id",
      referencedColumnName: "userId", // note: column in entity, not in table (which is user_id)
    }
  })
  public users: User[];

  public static async isUserInProject(userId: string, projectId: string) {
    const result = await getConnection()
      .createQueryBuilder(Project, "project")
      .innerJoin("project.users", "user")
      .where("user.userId = :userId", { userId })
      .andWhere("project.projectId = :projectId", { projectId })
      .select(["project.projectId"])
      .cache(`projects_users:${projectId}:${userId}`, 300000)
      .getOne();
    return !!result;
  }

  public async addUserByEmail(email: string) {
    // TODO
  }

  public async removeUserById(userId: string) {
    // TODO
    const cache = getConnection().queryResultCache;
    if (cache) {
      await cache.remove([`projects_users:${this.projectId}:${userId}`]);
    }
  }

}
