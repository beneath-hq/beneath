import { gql } from "apollo-server";
import { GraphQLResolveInfo } from "graphql";

import { User } from "../entities/User";

export const typeDefs = gql`
  extend type Query {
    user(username: String, userId: ID): User
  }

  type User {
    userId: ID!
    username: String
    name: String
    bio: String
    createdOn: Date
    updatedOn: Date
    projects: [Project]
  }
`;

export const resolvers = {
  Query: {
    user: async (root: any, args: any, ctx: any, info: GraphQLResolveInfo) => {
      return await User.findOne(args, { relations: ["projects"] });
    },
  },
};
