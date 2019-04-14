import { gql } from "apollo-server";
import { GraphQLResolveInfo } from "graphql";

import { User } from "../entities/User";
import { IApolloContext } from "../types";

export const typeDefs = gql`
  extend type Query {
    me: Me
    user(username: String, userId: ID): User
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
    projects: [Project]
  }
`;

export const resolvers = {
  Query: {
    me: async (root: any, args: any, ctx: IApolloContext, info: GraphQLResolveInfo) => {
      if (!ctx.user || !ctx.user.userId) {
        return null;
      }
      return await User.findOne({ userId: ctx.user.userId });
    },
    user: async (root: any, args: any, ctx: IApolloContext, info: GraphQLResolveInfo) => {
      return await User.findOne(args, { relations: ["projects"] });
    },
  },
};
