import { GraphQLScalarType } from "graphql";
import gql from "graphql-tag";
import { Kind } from "graphql/language";
import { merge } from "lodash";

import recordsSchema from "./local-schemas/records";

const schemas = [
  recordsSchema,
];

const schemasTypeDefs = schemas.map((schema) => schema.typeDefs);
const schemasResolvers = schemas.map((schema) => schema.resolvers);

const baseTypeDefs = gql`
  scalar JSON
`;

const baseResolvers = {
  Time: new GraphQLScalarType({
    name: "Time",
    description: "Time custom scalar type",
    parseValue(value) {
      return new Date(value); // value from the client
    },
    serialize(value) {
      return value.toISOString(); // value sent to the client
    },
    parseLiteral(ast) {
      if (ast.kind === Kind.INT) {
        return new Date(ast.value);
      } else if (ast.kind === Kind.STRING) {
        return new Date(ast.value);
      }
      return null;
    },
  }),
  UUID: new GraphQLScalarType({
    name: "UUID",
    description: "UUID custom scalar type",
    parseValue(value) {
      return value;
    },
    serialize(value) {
      return value;
    },
    parseLiteral(ast) {
      if (ast.kind === Kind.STRING) {
        return ast.value;
      }
      return null;
    },
  }),
  JSON: new GraphQLScalarType({
    name: "JSON",
    description: "JSON custom scalar type",
    parseValue(value) {
      return JSON.parse(value);
    },
    serialize(value) {
      return JSON.stringify(value);
    },
    parseLiteral(ast) {
      if (ast.kind === Kind.STRING) {
        return JSON.parse(ast.value);
      }
      return null;
    },
  }),
};

export const typeDefs = [baseTypeDefs, ...schemasTypeDefs];
export const resolvers = merge(baseResolvers, ...schemasResolvers);
