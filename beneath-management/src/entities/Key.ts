import { IsEmpty, ValidateIf } from "class-validator";
import crypto from "crypto";
import { getConnection } from "typeorm";
import uuidv4 from "uuid/v4"; // the secure random uuid

import {
  BaseEntity, Column, CreateDateColumn, Entity, JoinColumn,
  ManyToOne, PrimaryGeneratedColumn, RelationId, UpdateDateColumn,
} from "typeorm";

import { Project } from "./Project";
import { User } from "./User";

export enum KeyRole {
    Readonly = "r",
    Readwrite = "rw",
    Manage = "m",
}

@Entity("keys")
export class Key extends BaseEntity {

  @PrimaryGeneratedColumn("uuid", { name: "key_id" })
  public keyId: string;

  @Column({ length: 32, nullable: true })
  public description: string;

  @Column({ length: 4 })
  public prefix: string;

  @Column({ length: 64, unique: true })
  public hashedKey: string;

  @Column({ nullable: false })
  public role: KeyRole;

  @RelationId((key: Key) => key.user)
  @ValidateIf((key) => !!key.projectId)
  @IsEmpty()
  public userId: string;

  @RelationId((key: Key) => key.project)
  @ValidateIf((key) => !!key.userId)
  @IsEmpty()
  public projectId: string;

  @ManyToOne((type) => User, (user) => user.keys, { onDelete: "CASCADE" })
  @JoinColumn({ name: "user_id" })
  public user: User;

  @ManyToOne((type) => Project, (project) => project.keys, { onDelete: "CASCADE" })
  @JoinColumn({ name: "project_id" })
  public project: Project;

  @CreateDateColumn({ name: "created_on" })
  public createdOn: Date;

  @UpdateDateColumn({ name: "updated_on" })
  public updatedOn: Date;

  // transient property, only set after create
  public keyString: string;

  public static generateKey() {
    return (uuidv4() + uuidv4()).replace(/-/g, "");
  }

  public static hashKey(key) {
    return crypto.createHash("sha256").update(key).digest("hex");
  }

  public static newKey() {
    const key = new Key();
    key.keyString = Key.generateKey();
    key.hashedKey = Key.hashKey(key.keyString);
    key.prefix = key.keyString.slice(0, 4);
    return key;
  }

  public static async issueUserKey(userId: string, role: KeyRole, description: string) {
    const key = this.newKey();
    key.description = description;
    key.role = role;
    key.user = new User();
    key.user.userId = userId;
    await key.save();
    return key;
  }

  public static async issueProjectKey(projectId: string, role: KeyRole, description: string) {
    const key = this.newKey();
    key.description = description;
    key.role = role;
    key.project = new Project();
    key.project.projectId = projectId;
    await key.save();
    return key;
  }

  public static async authenticateKey(keyString: string): Promise<Key> {
    const hashedKey = this.hashKey(keyString);
    const key: Key = await this.findOne({ hashedKey }, {
      cache: { id: `key:${hashedKey}`, milliseconds: 3600000 }
    });
    if (!key) {
      return null;
    }
    return key;
  }

  public async revoke() {
    await this.remove();
    const cache = getConnection().queryResultCache;
    if (cache) {
      await cache.remove([`key:${this.hashedKey}`]);
    }
  }

}
