import { gql, UserInputError } from "apollo-server";
import { GraphQLResolveInfo } from "graphql";

import { Project } from "../entities/Project";
import { User } from "../entities/User";
import { IApolloContext } from "../types";

export const typeDefs = gql`
  extend type Query {
    project(name: String, projectId: ID): Project
  }

  extend type Mutation {
    addUserToProject(email: String!, projectId: ID!): User
    removeUserFromProject(userId: ID!, projectId: ID!): Boolean
  }

  type Project {
    projectId: ID!
    name: String
    displayName: String
    site: String
    description: String
    createdOn: Date
    updatedOn: Date
    users: [User]
  }
`;

export const resolvers = {
  Query: {
    project: async (root: any, args: any, ctx: IApolloContext, info: GraphQLResolveInfo) => {
      const project = await Project.findOne(args, { relations: ["users"] });
      await ctx.auth.requireCanReadProject(project.projectId);
      return project;
    },
  },
  Mutation: {
    addUserToProject: async (root: any, args: any, ctx: IApolloContext, info: GraphQLResolveInfo) => {
      const { projectId, email } = args;
      await ctx.auth.requireCanEditProject(projectId);

      const user = await User.findOneByEmail(email);
      if (!user) {
        throw new UserInputError("User not found");
      }

      const project = await Project.findOne({ projectId }, { relations: ["users"] });
      if (project.users.some(({ userId }) => userId === user.userId)) {
        throw new UserInputError("User already in project");
      }

      await project.addUser(user);
      return user;
    },
    removeUserFromProject: async (root: any, args: any, ctx: IApolloContext, info: GraphQLResolveInfo) => {
      const { projectId, userId } = args;
      await ctx.auth.requireCanEditProject(projectId);

      const project = await Project.findOne({ projectId }, { relations: ["users"] });
      if (project.users.length > 1) {
        await project.removeUserById(userId);
        return true;
      } else {
        return false;
      }
    },
  },
};
