import { IsEmpty, ValidateIf } from "class-validator";
import crypto from "crypto";
import { getConnection } from "typeorm";
import uuidv4 from "uuid/v4"; // the secure random uuid

import {
  BaseEntity, Column, CreateDateColumn, Entity, Index, JoinColumn,
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

  @Column({ length: 4, nullable: false })
  public prefix: string;

  @Column({ length: 44, name: "hashed_key", nullable: false })
  @Index("IDX_UQ_KEYS_HASHED_KEY", { unique: true })
  public hashedKey: string;

  @Column({ nullable: false })
  public role: KeyRole;

  @ManyToOne((type) => User, (user) => user.keys, { nullable: true, onDelete: "CASCADE" })
  @JoinColumn({ name: "user_id" })
  @ValidateIf((key) => !!key.project || !!key.projectId)
  @IsEmpty()
  public user: User;

  @RelationId((key: Key) => key.user)
  public userId: string;

  @ManyToOne((type) => Project, (project) => project.keys, { nullable: true, onDelete: "CASCADE" })
  @JoinColumn({ name: "project_id" })
  @ValidateIf((key) => !!key.user || !!key.userId)
  @IsEmpty()
  public project: Project;

  @RelationId((key: Key) => key.project)
  public projectId: string;

  @CreateDateColumn({ name: "created_on" })
  public createdOn: Date;

  @UpdateDateColumn({ name: "updated_on" })
  public updatedOn: Date;

  // transient property, only set after create
  public keyString: string;

  public static generateKey() {
    const arrayBuffer = new Array();
    uuidv4(null, arrayBuffer, 0);
    uuidv4(null, arrayBuffer, 16);
    const buffer = new Buffer(arrayBuffer);
    return buffer.toString("base64");
  }

  public static hashKey(key: string) {
    return crypto.createHash("sha256").update(new Buffer(key, "base64")).digest().toString("base64");
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
