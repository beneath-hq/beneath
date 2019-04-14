import { IsEmail, IsFQDN, IsLowercase, Length } from "class-validator";
import {
  BaseEntity, Column, CreateDateColumn, Entity, JoinTable,
  ManyToMany, PrimaryGeneratedColumn, UpdateDateColumn,
} from "typeorm";

import { Project } from "./Project";

@Entity("users")
export class User extends BaseEntity {

  @PrimaryGeneratedColumn("uuid", { name: "user_id" })
  public userId: string;

  @Column({ length: 16, unique: true, nullable: true  })
  @IsLowercase()
  @Length(3, 16)
  public username: string;

  @Column({ length: 320, unique: true })
  @IsEmail()
  public email: string;

  @Column({ length: 50 })
  @Length(4, 50)
  public name: string;

  @Column({ length: 256, nullable: true  })
  public bio: string;

  @Column({ length: 256, name: "photo_url", nullable: true })
  @IsFQDN()
  public photoUrl: string;

  @Column({ length: 256, name: "google_id", unique: true, nullable: true })
  public googleId: string;

  @Column({ length: 256, name: "github_id", unique: true, nullable: true })
  public githubId: string;

  @CreateDateColumn({ name: "created_on" })
  public createdOn: Date;

  @UpdateDateColumn({ name: "updated_on" })
  public updatedOn: Date;

  @ManyToMany((type) => Project, (project) => project.users)
  public projects: Project[];

  public static async createOrUpdate({ githubId, googleId, email, name, photoUrl }) {
    let user = null;
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
    }

    user.githubId = user.githubId || githubId;
    user.googleId = user.googleId || googleId;
    user.email = email;
    user.name = name;
    user.photoUrl = photoUrl;

    await user.save();
    return user;
  }
}
