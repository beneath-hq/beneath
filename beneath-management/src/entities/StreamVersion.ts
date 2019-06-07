import { IsNotEmpty, Length, Matches } from "class-validator";

import {
  BaseEntity, CreateDateColumn, Entity, JoinColumn, ManyToOne, 
  PrimaryGeneratedColumn, RelationId, UpdateDateColumn,
} from "typeorm";

import { Stream } from "./Stream";

@Entity("stream_versions")
export class StreamVersion extends BaseEntity {

  @PrimaryGeneratedColumn("uuid", { name: "stream_version_id" })
  public streamVersionId: string;

  @ManyToOne((type) => Stream, (stream) => stream.streamVersions, { nullable: false, onDelete: "CASCADE" })
  @JoinColumn({ name: "stream_id" })
  @IsNotEmpty()
  public stream: Stream;

  @RelationId((streamVersion: StreamVersion) => streamVersion.stream)
  public streamId: string;

  @CreateDateColumn({ name: "created_on" })
  public createdOn: Date;

  @UpdateDateColumn({ name: "updated_on" })
  public updatedOn: Date;

}
