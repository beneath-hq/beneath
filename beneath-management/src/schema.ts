import { gql } from "apollo-server";
import { GraphQLScalarType } from "graphql";
import { Kind } from "graphql/language";
import { merge } from "lodash";

const schemas = [
  "./schemas/keys",
  "./schemas/projects",
  "./schemas/users",
].map(require);

const schemasTypeDefs = schemas.map((schema) => schema.typeDefs);
const schemasResolvers = schemas.map((schema) => schema.resolvers);

const baseTypeDefs = gql`
  type Query {
    _empty: String
    ping: String!
  }
  type Mutation {
    _empty: String
  }
  type Subscription {
    _empty: String
  }
  scalar Date
`;

const baseResolvers = {
  Query: {
    ping: () => {
      return "pong";
    }
  },
  Date: new GraphQLScalarType({
    name: "Date",
    description: "Date custom scalar type",
    parseValue(value) {
      return new Date(value); // value from the client
    },
    serialize(value) {
      return value.getTime(); // value sent to the client
    },
    parseLiteral(ast) {
      if (ast.kind === Kind.INT) {
        return new Date(ast.value) // ast value is always in string format
      }
      return null;
    },
  })
};

export const typeDefs = [baseTypeDefs, ...schemasTypeDefs];
export const resolvers = merge(baseResolvers, ...schemasResolvers);
