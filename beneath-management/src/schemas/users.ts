import { gql } from "apollo-server";
import { GraphQLResolveInfo } from "graphql";

import { User } from "../entities/User";
import { requireValidates } from "../lib/guards";
import { IApolloContext } from "../types";

export const typeDefs = gql`
  extend type Query {
    me: Me
    user(userId: ID!): User
  }

  extend type Mutation {
    updateMe(name: String, bio: String): Me
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
    user: User!
    email: String
    updatedOn: Date
    keys: [Key]
  }
`;

const userToMe = (user) => {
  return {
    userId: user.userId,
    user: {
      userId: user.userId,
      username: user.username,
      name: user.name,
      bio: user.bio,
      photoUrl: user.photoUrl,
      createdOn: user.createdOn,
      projects: user.projects,
    },
    email: user.email,
    updatedOn: user.updatedOn,
    keys: user.keys,
  };
};

export const resolvers = {
  Query: {
    me: async (root: any, args: any, ctx: IApolloContext, info: GraphQLResolveInfo) => {
      ctx.auth.requirePersonalUser();
      const user = await User.findOne({ userId: ctx.auth.getUserId() }, { relations: ["keys", "projects"] });
      return userToMe(user);
    },
    user: async (root: any, args: any, ctx: IApolloContext, info: GraphQLResolveInfo) => {
      let userId = args.userId;
      if (userId === "me") {
        ctx.auth.requirePersonalUser();
        userId = ctx.auth.getUserId();
      }
      return await User.findOne({ userId }, { relations: ["projects"] });
    },
  },
  Mutation: {
    updateMe: async (root: any, args: any, ctx: IApolloContext, info: GraphQLResolveInfo) => {
      ctx.auth.requirePersonalUser();
      const user = await User.findOne({ userId: ctx.auth.getUserId() });
      if (args.name !== undefined) {
        user.name = args.name;
      }
      if (args.bio !== undefined) {
        user.bio = args.bio;
      }
      await requireValidates(user);
      await user.save();
      return userToMe(user);
    },
  },
};
