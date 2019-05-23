import { IsEmail, IsUrl, IsLowercase, Length, Matches } from "class-validator";
import {
  BaseEntity, Column, CreateDateColumn, Entity, getConnection, Index,
  ManyToMany, OneToMany, PrimaryGeneratedColumn, UpdateDateColumn,
} from "typeorm";

import logger from "../lib/logger";
import { Key } from "./Key";
import { Project } from "./Project";

@Entity("users")
export class User extends BaseEntity {

  @PrimaryGeneratedColumn("uuid", { name: "user_id" })
  public userId: string;

  @Column({ length: 16, nullable: true  })
  @Index("IDX_UQ_USERS_USERNAME", { unique: true })
  @IsLowercase()
  @Length(3, 16)
  @Matches(/^[_a-z][_\-a-z0-9]*$/)
  public username: string;

  @Column({ length: 320, nullable: false })
  @Index("IDX_UQ_USERS_EMAIL", { unique: true })
  @IsEmail()
  public email: string;

  @Column({ length: 50, nullable: false })
  @Length(4, 50)
  public name: string;

  @Column({ length: 255, nullable: true })
  @Length(0, 255)
  public bio: string;

  @Column({ length: 255, name: "photo_url", nullable: true })
  @IsUrl()
  @Length(0, 255)
  public photoUrl: string;

  @Column({ length: 256, name: "google_id", nullable: true })
  @Index("IDX_UQ_USERS_GOOGLE_ID", { unique: true })
  public googleId: string;

  @Column({ length: 256, name: "github_id", nullable: true })
  @Index("IDX_UQ_USERS_GITHUB_ID", { unique: true })
  public githubId: string;

  @CreateDateColumn({ name: "created_on" })
  public createdOn: Date;

  @UpdateDateColumn({ name: "updated_on" })
  public updatedOn: Date;

  @ManyToMany((type) => Project, (project) => project.users)
  public projects: Project[];

  @OneToMany((type) => Key, (key) => key.user)
  public keys: Key[];

  public static async findOneByEmail(email: string) {
    return await getConnection()
      .createQueryBuilder(User, "user")
      .where("lower(user.email) = lower(:email)", { email })
      .getOne();
  }

  public static async createOrUpdate({ githubId, googleId, email, name, photoUrl }) {
    let user = null;
    let created = false;
    if (githubId) {
      user = await User.findOne({ githubId });
    } else if (googleId) {
      user = await User.findOne({ googleId });
    }
    if (!user) {
      user = await User.findOne({ email });
    }
    if (!user) {
      user = new User();
      created = true;
    }

    user.githubId = user.githubId || githubId;
    user.googleId = user.googleId || googleId;
    user.email = email;
    user.name = name;
    user.photoUrl = photoUrl;

    await user.save();
    if (created) {
      logger.info(`Created userId <${user.userId}>`);
    } else {
      logger.info(`Updated userId <${user.userId}>`);
    }
    return user;
  }

}
