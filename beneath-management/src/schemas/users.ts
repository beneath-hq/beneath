import { AuthenticationError, ForbiddenError, gql } from "apollo-server";
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
    updateMe(name: String, bio: String): Me
    issueKey(description: String!, readonly: Boolean!): NewKey!
    revokeKey(keyId: ID!): Boolean
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
      if (ctx.user.anonymous) {
        throw new AuthenticationError("Must be authenticated");
      } else if (ctx.user.key.role !== "personal") {
        throw new ForbiddenError("Only permitted with personal login");
      }
      return await User.findOne({ userId: ctx.user.key.userId }, { relations: ["keys", "projects"] });
    },
    user: async (root: any, args: any, ctx: IApolloContext, info: GraphQLResolveInfo) => {
      return await User.findOne(args, { relations: ["projects"] });
    },
  },
  Mutation: {
    updateMe: async (root: any, args: any, ctx: IApolloContext, info: GraphQLResolveInfo) => {
      if (ctx.user.anonymous) {
        throw new AuthenticationError("Must be authenticated");
      } else if (ctx.user.key.role !== "personal") {
        throw new ForbiddenError("Only permitted with personal login");
      }
      const user = await User.findOne({ userId: ctx.user.key.userId });
      if (args.name) {
        user.name = args.name;
      }
      if (args.bio) {
        user.bio = args.bio;
      }
      await user.save();
      return user;
    },
    issueKey: async (root: any, args: any, ctx: IApolloContext, info: GraphQLResolveInfo) => {
      if (ctx.user.anonymous) {
        throw new AuthenticationError("Must be authenticated");
      } else if (ctx.user.key.role !== "personal") {
        throw new ForbiddenError("Only permitted with personal login");
      }
      const role: KeyRole = args.readonly ? "readonly" : "readwrite";
      const key = await Key.issueKey({ description: args.description, role, userId: ctx.user.key.userId });
      return {
        key,
        keyString: key.keyString,
      };
    },
    revokeKey: async (root: any, args: any, ctx: IApolloContext, info: GraphQLResolveInfo) => {
      if (ctx.user.anonymous) {
        throw new AuthenticationError("Must be authenticated");
      } else if (ctx.user.key.role !== "personal") {
        throw new ForbiddenError("Only permitted with personal login");
      }
      const key = await Key.findOneOrFail({ keyId: args.keyId, userId: ctx.user.key.userId });
      await key.remove();
      return true;
    },
  },
};
