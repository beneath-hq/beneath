import { IsUrl, IsLowercase, Length, Matches, ValidateIf } from "class-validator";
import {
  BaseEntity, Column, CreateDateColumn, Entity, getConnection, Index, 
  JoinTable, ManyToMany, OneToMany, PrimaryGeneratedColumn, UpdateDateColumn,
} from "typeorm";

import { Key } from "./Key";
import { User } from "./User";
import { Stream } from "./Stream";

@Entity("projects")
export class Project extends BaseEntity {

  @PrimaryGeneratedColumn("uuid", { name: "project_id" })
  public projectId: string;

  @Column({ length: 16, nullable: false })
  @Index("IDX_UQ_PROJECTS_NAME", { unique: true })
  @IsLowercase()
  @Length(3, 16)
  @Matches(/^[_a-z][_\-a-z0-9]*$/)
  public name: string;

  @Column({ length: 40, nullable: false, name: "display_name" })
  @Length(3, 40)
  public displayName: string;

  @Column({ length: 255, nullable: true })
  @ValidateIf((project) => (!!project.site))
  @IsUrl()
  public site: string;

  @Column({ length: 255, nullable: true })
  @Length(0, 255)
  public description: string;

  @Column({ length: 255, name: "photo_url", nullable: true })
  @ValidateIf((project) => (!!project.photoUrl))
  @IsUrl()
  @Length(0, 255)
  public photoUrl: string;

  @CreateDateColumn({ name: "created_on" })
  public createdOn: Date;

  @UpdateDateColumn({ name: "updated_on" })
  public updatedOn: Date;

  @OneToMany((type) => Key, (key) => key.project)
  public keys: Key[];
  
  @OneToMany((type) => Stream, (stream) => stream.project)
  public streams: Stream[];

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

  public async addUser(user: User) {
    this.users.push(user);
    await this.save();
  }

  public async removeUserById(userId: string) {
    await Project.createQueryBuilder()
      .relation("users")
      .of({ projectId: this.projectId })
      .remove({ userId });

    const cache = getConnection().queryResultCache;
    if (cache) {
      await cache.remove([`projects_users:${this.projectId}:${userId}`]);
    }
  }

}
