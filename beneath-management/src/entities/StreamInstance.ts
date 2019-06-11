import { IsNotEmpty, Length, Matches } from "class-validator";

import {
  BaseEntity, CreateDateColumn, Entity, JoinColumn, ManyToOne, 
  PrimaryGeneratedColumn, RelationId, UpdateDateColumn,
} from "typeorm";

import { Stream } from "./Stream";

@Entity("stream_instances")
export class StreamInstance extends BaseEntity {

  @PrimaryGeneratedColumn("uuid", { name: "stream_instance_id" })
  public streamInstanceId: string;

  @ManyToOne((type) => Stream, (stream) => stream.streamInstances, { nullable: false, onDelete: "CASCADE" })
  @JoinColumn({ name: "stream_id" })
  @IsNotEmpty()
  public stream: Stream;

  @RelationId((streamInstance: StreamInstance) => streamInstance.stream)
  public streamId: string;

  @CreateDateColumn({ name: "created_on" })
  public createdOn: Date;

  @UpdateDateColumn({ name: "updated_on" })
  public updatedOn: Date;

}
