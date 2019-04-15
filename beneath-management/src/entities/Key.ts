import crypto from "crypto";
import uuidv4 from "uuid/v4"; // the secure random uuid

import {
  BaseEntity, Column, CreateDateColumn, Entity, JoinColumn,
  ManyToOne, PrimaryGeneratedColumn, RelationId, UpdateDateColumn,
} from "typeorm";

import { KeyRole } from "../types";
import { User } from "./User";

@Entity("keys")
export class Key extends BaseEntity {

  @PrimaryGeneratedColumn("uuid", { name: "key_id" })
  public keyId: string;

  @Column({ length: 32, nullable: true })
  public name: string;

  @ManyToOne((type) => User, (user) => user.keys, { nullable: false, onDelete: "CASCADE" })
  @JoinColumn({ name: "user_id" })
  public user: User;

  @RelationId((key: Key) => key.user)
  public userId: string;

  @Column({ length: 4 })
  public prefix: string;

  @Column({ length: 64, unique: true })
  public hashedKey: string;

  @Column({ nullable: false })
  public role: string;

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

  public static async issueKey({ name, role, userId }: { name: string, role: KeyRole, userId: string }) {
    const keyString = Key.generateKey();
    const key = new Key();
    key.name = name;
    key.user = new User();
    key.user.userId = userId;
    key.prefix = keyString.slice(0, 4);
    key.hashedKey = Key.hashKey(keyString);
    key.role = role;
    await key.save();
    key.keyString = keyString;
    return key;
  }

  public static async authenticateKey(keyString): Promise<Key> {
    const hashedKey = this.hashKey(keyString);
    const key: Key = await this.findOne({ hashedKey }, {
      cache: { id: `key:${hashedKey}`, milliseconds: 3600000 }
    });

    if (!key) {
      return null;
    }

    return key;
  }

}
