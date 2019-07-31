import gql from "graphql-tag";

import connection from "../../lib/connection";
import { GET_TOKEN } from "../queries/local/token";

export const typeDefs = gql`
  extend type Query {
    records(
      projectName: String!,
      streamName: String!,
      keyFields: [String!]!,
      where: JSON,
      limit: Int!,
    ): [Record!]!
  }

  type Record {
    recordID: ID!
    data: JSON!
    sequenceNumber: String!
  }
`;

export const resolvers = {
  Query: {
    records: async (_: any, { projectName, streamName, keyFields, where, limit }: any, { cache }: any) => {
      // build url with limit and where
      let url = `${connection.GATEWAY_URL}/projects/${projectName}/streams/${streamName}`;
      url += `?limit=${limit}`;
      if (where) {
        url += `&where=${JSON.stringify(where)}`;
      }

      // build headers with authorization
      const headers: any = { "Content-Type": "application/json" };
      const { token } = cache.readQuery({ query: GET_TOKEN });
      if (token) {
        headers.Authorization = `Bearer ${token}`;
      }

      // fetch
      const res = await fetch(url, { headers });
      const json = await res.json();

      // build records objects
      return json.map((row: any) => {
        return {
          __typename: "Record",
          recordID: makeUniqueIdentifier(keyFields, row),
          data: row,
          sequenceNumber: row["@meta"].sequence_number,
        };
      });
    },
  },
};

export default {
  typeDefs,
  resolvers,
};

const makeUniqueIdentifier = (keyFields: string[], data: any) => {
  return keyFields.reduce((prev, curr) => `${data[prev]}-${data[curr]}`, "");
};
