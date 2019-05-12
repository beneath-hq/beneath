import { AuthenticationError, ForbiddenError, gql } from "apollo-server";
import { GraphQLResolveInfo } from "graphql";

import { Key } from "../entities/Key";
import { IApolloContext, KeyRole } from "../types";

export const typeDefs = gql`
  # extend type Query {
  # }

  extend type Mutation {
    issueKey(description: String!, readonly: Boolean!): NewKey!
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
  },
  Mutation: {
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
