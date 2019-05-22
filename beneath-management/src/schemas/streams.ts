import { gql, UserInputError } from "apollo-server";
import { GraphQLResolveInfo } from "graphql";

import { Project } from "../entities/Project";
import { Stream, SchemaType } from "../entities/Stream";
import { NotFoundError } from "../lib/errors";
import { requireValidates } from "../lib/guards";
import { IApolloContext } from "../types";

export const typeDefs = gql`
  extend type Query {
    stream(name: String!, projectName: String!): Stream
  }

  extend type Mutation {
    createExternalStream(
      projectId: ID!,
      name: String!,
      description: String!,
      avroSchema: String!,
      batch: Boolean!,
      manual: Boolean!
    ): Stream!
  }

  type Stream {
    streamId: ID!
    name: String!
    description: String
    schema: String
    schemaType: String
    batch: Boolean
    external: Boolean
    manual: Boolean
    project: Project
    createdOn: Date
    updatedOn: Date
  }
`;

export const resolvers = {
  Query: {
    stream: async (root: any, args: any, ctx: IApolloContext, info: GraphQLResolveInfo) => {
      const stream = await Stream.findOneByNameAndProject(args.name, args.projectName);
      if (!stream) {
        throw new NotFoundError("Stream not found");
      }
      await ctx.auth.requireCanReadProject(stream.projectId);
      return stream;
    },
  },
  Mutation: {
    createExternalStream: async (root: any, args: any, ctx: IApolloContext, info: GraphQLResolveInfo) => {
      await ctx.auth.requireCanEditProject(args.projectId);
      const stream = new Stream();
      stream.project = await Project.findOne({ projectId: args.projectId });
      stream.name = args.name;
      stream.description = args.description;
      stream.schema = args.avroSchema;
      stream.schemaType = SchemaType.Avro;
      stream.compiledAvroSchema = args.avroSchema;
      stream.batch = args.batch;
      stream.external = true;
      stream.manual = args.manual;
      console.log(stream);
      await requireValidates(stream);
      await stream.save();
      return stream;
    },
  },
};
