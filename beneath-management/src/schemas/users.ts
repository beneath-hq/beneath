import { ForbiddenError, gql } from "apollo-server";
import { GraphQLResolveInfo } from "graphql";

import { User } from "../entities/User";
import { IApolloContext } from "../types";

export const typeDefs = gql`
  extend type Query {
    me: Me
    user(username: String, userId: ID): User
  }

  extend type Mutation {
    issueKey(name: String!, modifyScope: Boolean!): NewKey!
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
    name: String
    prefix: String
    modifyScope: Boolean
    revoked: Boolean
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
      if (ctx.user.kind !== "session" && ctx.user.kind !== "secret") { // TOOD: Only session in production
        return null;
      }
      return await User.findOne({ userId: ctx.user.userId }, { relations: ["keys", "projects"] });
    },
    user: async (root: any, args: any, ctx: IApolloContext, info: GraphQLResolveInfo) => {
      return await User.findOne(args, { relations: ["projects"] });
    },
  },
  Mutation: {
    issueKey: async (root: any, args: any, ctx: IApolloContext, info: GraphQLResolveInfo) => {
      if (ctx.user.kind !== "session") {
        throw new ForbiddenError("Only logged-in users can issue keys");
      }
      const user = new User();
      user.userId = ctx.user.userId;
      const key = await user.issueKey({ name: args.name, modifyScope: args.modifyScope });
      return {
        key,
        keyString: key.keyString,
      };
    },
  },
};
