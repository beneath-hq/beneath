import crypto from "crypto";
import uuidv4 from "uuid/v4"; // the secure random uuid

import {
  BaseEntity, Column, CreateDateColumn, Entity, JoinColumn,
  ManyToOne, PrimaryGeneratedColumn, UpdateDateColumn,
} from "typeorm";

import logger from "../lib/logger";
import { User } from "./User";

@Entity("keys")
export class Key extends BaseEntity {

  @PrimaryGeneratedColumn("uuid", { name: "key_id" })
  public keyId: string;

  @Column({ length: 32, nullable: true })
  public name: string;

  @ManyToOne((type) => User, (user) => user.keys)
  @JoinColumn({ name: "user_id" })
  public user: User;

  @Column({ length: 8 })
  public prefix: string;

  @Column({ length: 64, unique: true })
  public hashedKey: string;

  @Column({ nullable: false, default: false })
  public modifyScope: boolean;

  @Column({ nullable: false, default: false })
  public revoked: boolean;

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

  public static async authenticateKey(keyString) {
    const hashedKey = this.hashKey(keyString);
    const key: Key = await this.findOne({ hashedKey }, { cache: { id: "keys", milliseconds: 60000 } });

    if (!key || key.revoked) {
      return null;
    }

    const scopes = [];
    if (key.modifyScope) {
      scopes.push("modify");
    }

    return {
      scopes,
      userId: key.user.userId,
    };
  }

}
