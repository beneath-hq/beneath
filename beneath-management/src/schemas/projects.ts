import { gql, UserInputError } from "apollo-server";
import { GraphQLResolveInfo } from "graphql";

import { Project } from "../entities/Project";
import { User } from "../entities/User";
import { NotFoundError } from "../lib/errors";
import { requireValidates } from "../lib/guards";
import { IApolloContext } from "../types";

export const typeDefs = gql`
  extend type Query {
    project(name: String, projectId: ID): Project
  }

  extend type Mutation {
    createProject(name: String!, displayName: String!, site: String, description: String, photoUrl: String): Project!
    updateProject(projectId: ID!, displayName: String, site: String, description: String, photoUrl: String): Project!
    addUserToProject(email: String!, projectId: ID!): User
    removeUserFromProject(userId: ID!, projectId: ID!): Boolean
  }

  type Project {
    projectId: ID!
    name: String
    displayName: String
    site: String
    description: String
    photoUrl: String
    createdOn: Date
    updatedOn: Date
    users: [User]
  }
`;

export const resolvers = {
  Query: {
    project: async (root: any, args: any, ctx: IApolloContext, info: GraphQLResolveInfo) => {
      const project = await Project.findOne(args, { relations: ["users"] });
      if (!project) {
        throw new NotFoundError("Project not found");
      }
      await ctx.auth.requireCanReadProject(project.projectId);
      return project;
    },
  },
  Mutation: {
    createProject: async (root: any, args: any, ctx: IApolloContext, info: GraphQLResolveInfo) => {
    },
    updateProject: async (root: any, args: any, ctx: IApolloContext, info: GraphQLResolveInfo) => {
      const { projectId, displayName, site, description, photoUrl } = args;
      ctx.auth.requireCanEditProject(projectId);
      const project = await Project.findOne({ projectId });
      if (displayName !== undefined) {
        project.displayName = displayName;
      }
      if (site !== undefined) {
        project.site = site;
      }
      if (description !== undefined) {
        project.description = description;
      }
      if (photoUrl !== undefined) {
        project.photoUrl = photoUrl;
      }
      await requireValidates(project);
      await project.save();
      return project;
    },
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
