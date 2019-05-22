import { IsNotEmpty, Matches, ValidateIf } from "class-validator";

import {
  BaseEntity, Column, CreateDateColumn, Entity, getConnection, Index, 
  JoinColumn, ManyToOne, PrimaryGeneratedColumn, RelationId, UpdateDateColumn,
} from "typeorm";

import { Project } from "./Project";
import { IsAvroSchema } from "../lib/validators";

export enum SchemaType {
  Avro = "av",
  // in the future: gql sdl, bq, sql
}

@Entity("streams")
@Index("STREAM_PROJECT_NAME_UNIQUE", ["project", "name"], { unique: true })
export class Stream extends BaseEntity {

  @PrimaryGeneratedColumn("uuid", { name: "stream_id" })
  public streamId: string;

  @Column({ length: 40, nullable: true })
  @Matches(/^[_a-z][_\-a-z0-9]*$/)
  public name: string;

  @Column({ length: 256, nullable: true })
  public description: string;

  @Column()
  public schema: string;

  @Column({ name: "schema_type" })
  public schemaType: SchemaType;

  @Column({ type: "json" })
  @IsAvroSchema()
  public compiledAvroSchema: any;

  @Column()
  public batch: boolean;

  @Column()
  public manual: boolean;

  @Column()
  public external: boolean;

  @RelationId((stream: Stream) => stream.project)
  // @Column({ name: "project_id" })
  public projectId: string;
  
  @ManyToOne((type) => Project, (project) => project.streams, { nullable: false, onDelete: "SET NULL" })
  @JoinColumn({ name: "project_id" })
  @IsNotEmpty()
  public project: Project;

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
