import { IsEmail, IsLowercase, Length } from "class-validator";
import {
  BaseEntity, Column, CreateDateColumn, Entity, JoinTable,
  ManyToMany, PrimaryGeneratedColumn, UpdateDateColumn,
} from "typeorm";

import { Project } from "./Project";

@Entity("users")
export class User extends BaseEntity {

  @PrimaryGeneratedColumn("uuid", { name: "user_id" })
  public userId: string;

  @Column({ length: 16, unique: true })
  @IsLowercase()
  @Length(3, 16)
  public username: string;

  @Column({ length: 320, unique: true })
  @IsEmail()
  public email: string;

  @Column({ length: 50 })
  @Length(4, 50)
  public name: string;

  @Column({ length: 256 })
  public bio: string;

  @CreateDateColumn({ name: "created_on" })
  public createdOn: Date;

  @UpdateDateColumn({ name: "updated_on" })
  public updatedOn: Date;

  @ManyToMany((type) => Project, (project) => project.users)
  public projects: Project[];

}
