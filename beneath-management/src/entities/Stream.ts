import { IsNotEmpty, Length, Matches } from "class-validator";

import {
  BaseEntity, Column, CreateDateColumn, Entity, getConnection, Index, JoinColumn, 
  ManyToOne, OneToMany, OneToOne, PrimaryGeneratedColumn, RelationId, UpdateDateColumn,
} from "typeorm";

import { Project } from "./Project";
import { StreamInstance } from "./StreamInstance";
import { IsAvroSchema } from "../lib/validators";

export enum SchemaType {
  Avro = "av",
  // in the future: gql sdl, bq, sql
}

@Entity("streams")
@Index("IDX_UQ_STREAMS_NAME_PROJECT", ["project", "name"], { unique: true })
export class Stream extends BaseEntity {

  @PrimaryGeneratedColumn("uuid", { name: "stream_id" })
  public streamId: string;

  @Column({ length: 40, nullable: false })
  // See unique index definition above class
  @Matches(/^[_a-z][_\-a-z0-9]*$/)
  @Length(1, 40)
  public name: string;

  @Column({ length: 255, nullable: true })
  @Length(0, 255)
  public description: string;

  @Column({ nullable: false })
  public schema: string;

  @Column({ name: "schema_type", nullable: false })
  public schemaType: SchemaType;

  @Column({ nullable: false, type: "json" })
  @IsAvroSchema()
  public compiledAvroSchema: any;

  @Column({ nullable: false })
  public batch: boolean;

  @Column({ nullable: false })
  public manual: boolean;

  @Column({ nullable: false })
  public external: boolean;
  
  @ManyToOne((type) => Project, (project) => project.streams, { nullable: false, onDelete: "SET NULL" })
  @JoinColumn({ name: "project_id" })
  @IsNotEmpty()
  public project: Project;

  @RelationId((stream: Stream) => stream.project)
  public projectId: string;

  @OneToMany((type) => StreamInstance, (streamInstance) => streamInstance.stream)
  public streamInstances: StreamInstance[];

  @OneToOne((type) => StreamInstance, { nullable: true, onDelete: "SET NULL" })
  @JoinColumn({ name: "current_stream_instance_id" })
  public currentStreamInstance: StreamInstance;

  @CreateDateColumn({ name: "created_on" })
  public createdOn: Date;

  @UpdateDateColumn({ name: "updated_on" })
  public updatedOn: Date;

  public static async findOneByNameAndProject(name: string, projectName: string) {
    return await getConnection()
      .createQueryBuilder(Stream, "stream")
      .innerJoinAndSelect("stream.project", "project")
      .where("stream.name = lower(:name)", { name })
      .andWhere("project.name = lower(:projectName)", { projectName })
      .getOne();
  }

  /**
   * TODO: In the future
   * - belongs to model
   * - dependencies (models)
   * - define where to sink (i.e. do not presume pubsub, bq, bt)
   */

}
