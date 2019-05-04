import { ForbiddenError, gql } from "apollo-server";
import { GraphQLResolveInfo } from "graphql";

import { Key } from "../entities/Key";
import { User } from "../entities/User";
import { IApolloContext, KeyRole } from "../types";

export const typeDefs = gql`
  extend type Query {
    me: Me
    user(username: String, userId: ID): User
  }

  extend type Mutation {
    issueKey(description: String!, readonly: Boolean!): NewKey!
  }

  type User {
    userId: ID!
    username: String
    name: String
    bio: String
    photoUrl: String
    createdOn: Date
    projects: [Project]
  }

  type Me {
    userId: ID!
    email: String
    username: String
    name: String
    bio: String
    photoUrl: String
    createdOn: Date
    updatedOn: Date
    keys: [Key]
    projects: [Project]
  }

  type Key {
    keyId: ID!
    description: String
    prefix: String
    role: String
    createdOn: Date
    updatedOn: Date
  }

  type NewKey {
    key: Key
    keyString: String
  }
`;

export const resolvers = {
  Query: {
    me: async (root: any, args: any, ctx: IApolloContext, info: GraphQLResolveInfo) => {
      if (ctx.user.anonymous || ctx.user.key.role !== "personal") {
        return null;
      }
      return await User.findOne({ userId: ctx.user.key.userId }, { relations: ["keys", "projects"] });
    },
    user: async (root: any, args: any, ctx: IApolloContext, info: GraphQLResolveInfo) => {
      return await User.findOne(args, { relations: ["projects"] });
    },
  },
  Mutation: {
    issueKey: async (root: any, args: any, ctx: IApolloContext, info: GraphQLResolveInfo) => {
      if (ctx.user.anonymous || ctx.user.key.role !== "personal") {
        throw new ForbiddenError("Only logged-in users can issue keys");
      }
      const role: KeyRole = args.readonly ? "readonly" : "readwrite";
      const key = await Key.issueKey({ description: args.description, role, userId: ctx.user.key.userId });
      return {
        key,
        keyString: key.keyString,
      };
    },
  },
};
