import { gql } from "apollo-server";
import { GraphQLResolveInfo } from "graphql";

import { Key, KeyRole } from "../entities/Key";
import { requireExclusiveArgs } from "../lib/guards";
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
      requireExclusiveArgs(args, ["projectId", "userId"]);

      const findConditions: any = {};
      if (args.userId) {
        ctx.auth.requireCanEditUser(args.userId);
        findConditions.user = { userId: args.userId };
      } else if (args.projectId) {
        await ctx.auth.requireCanEditProject(args.projectId);
        findConditions.project = { projectId: args.projectId };
      }

      return await Key.find(findConditions);
    },
  },
  Mutation: {
    issueKey: async (root: any, args: any, ctx: IApolloContext, info: GraphQLResolveInfo) => {
      requireExclusiveArgs(args, ["projectId", "userId"]);

      const role: KeyRole = args.readonly ? KeyRole.Readonly : KeyRole.Readwrite;

      let key = null;
      if (args.userId) {
        ctx.auth.requireCanEditUser(args.userId);
        key = await Key.issueUserKey(args.userId, role, args.description);
      } else if (args.projectId) {
        await ctx.auth.requireCanEditProject(args.projectId);
        key = await Key.issueProjectKey(args.projectId, role, args.description);
      }

      return {
        key,
        keyString: key.keyString,
      };
    },
    revokeKey: async (root: any, args: any, ctx: IApolloContext, info: GraphQLResolveInfo) => {
      const key = await Key.findOneOrFail({ keyId: args.keyId });

      if (key.userId) {
        ctx.auth.requireCanEditUser(key.userId);
      } else if (key.projectId) {
        await ctx.auth.requireCanEditProject(key.projectId);
      }

      await key.revoke();

      return true;
    },
  },
};
