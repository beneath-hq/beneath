import { AuthenticationError, ForbiddenError, gql } from "apollo-server";
import { GraphQLResolveInfo } from "graphql";

import { Key, KeyRole } from "../entities/Key";
import { ArgsError } from "../lib/errors";
import { IApolloContext } from "../types";

export const typeDefs = gql`
  extend type Query {
    keys(userId: ID, projectId: ID): [Key!]!
  }

  extend type Mutation {
    issueKey(userId: ID, projectId: ID, readonly: Boolean!, description: String): NewKey!
    revokeKey(keyId: ID!): Boolean
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
    keys: async (root: any, args: any, ctx: IApolloContext, info: GraphQLResolveInfo) => {
      if (ctx.user.anonymous) {
        throw new AuthenticationError("Must be authenticated");
      } else if (!ctx.user.key.userId || ctx.user.key.role !== KeyRole.Manage) {
        throw new ForbiddenError("Only permitted with manage login");
      }

      if (!!args.userId === !!args.projectId) {
        throw new ArgsError("Pass userId xor projectId");
      }

      // TODO: Check that ctx.user.key.userId has permissions (for project)

      const findConditions: any = {};
      if (args.userId) {
        findConditions.user = { userId: args.userId };
      } else if (args.projectId) {
        findConditions.project = { projectId: args.projectId };
      }

      return await Key.find(findConditions);
    },
  },
  Mutation: {
    issueKey: async (root: any, args: any, ctx: IApolloContext, info: GraphQLResolveInfo) => {
      if (ctx.user.anonymous) {
        throw new AuthenticationError("Must be authenticated");
      } else if (!ctx.user.key.userId || ctx.user.key.role !== KeyRole.Manage) {
        throw new ForbiddenError("Only permitted with manage login");
      }

      if (!!args.userId === !!args.projectId) {
        throw new ArgsError("Pass userId xor projectId");
      }

      if (args.userId && args.userId !== ctx.user.key.userId) {
        throw new ArgsError("Can only issue keys for yourself");
      }

      const role: KeyRole = args.readonly ? KeyRole.Readonly : KeyRole.Readwrite;
      let key = null;
      if (args.userId) {
        key = await Key.issueUserKey(args.userId, role, args.description);
      } else if (args.projectId) {
        // TODO: Check that ctx.user.key.userId has permissions (for project)
        key = await Key.issueProjectKey(args.projectId, role, args.description);
      }

      return {
        key,
        keyString: key.keyString,
      };
    },
    revokeKey: async (root: any, args: any, ctx: IApolloContext, info: GraphQLResolveInfo) => {
      if (ctx.user.anonymous) {
        throw new AuthenticationError("Must be authenticated");
      } else if (!ctx.user.key.userId || ctx.user.key.role !== KeyRole.Manage) {
        throw new ForbiddenError("Only permitted with personal login");
      }

      // TODO: Check that ctx.user.key.userId has permissions (user or project) for args.keyId

      const key = await Key.findOneOrFail({ keyId: args.keyId });
      await key.revoke();
      return true;
    },
  },
};
