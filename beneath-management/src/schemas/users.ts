import { AuthenticationError, ForbiddenError, gql } from "apollo-server";
import { GraphQLResolveInfo } from "graphql";

import { KeyRole } from "../entities/Key";
import { User } from "../entities/User";
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
      if (ctx.user.anonymous) {
        throw new AuthenticationError("Must be authenticated");
      } else if (!ctx.user.key.userId || ctx.user.key.role !== KeyRole.Manage) {
        throw new ForbiddenError("Only permitted with personal login");
      }
      const user = await User.findOne({ userId: ctx.user.key.userId }, { relations: ["keys", "projects"] });
      return userToMe(user);
    },
    user: async (root: any, args: any, ctx: IApolloContext, info: GraphQLResolveInfo) => {
      let userId = args.userId;
      if (userId === "me") {
        if (ctx.user.anonymous) {
          throw new AuthenticationError("Must be authenticated");
        } else if (!ctx.user.key.userId || ctx.user.key.role !== KeyRole.Manage) {
          throw new ForbiddenError("Only permitted with personal login");
        }
        userId = ctx.user.key.userId;
      }
      return await User.findOne({ userId }, { relations: ["projects"] });
    },
  },
  Mutation: {
    updateMe: async (root: any, args: any, ctx: IApolloContext, info: GraphQLResolveInfo) => {
      if (ctx.user.anonymous) {
        throw new AuthenticationError("Must be authenticated");
      } else if (!ctx.user.key.userId || ctx.user.key.role !== KeyRole.Manage) {
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
      return userToMe(user);
    },
  },
};
