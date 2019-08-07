import { ApolloCache } from "apollo-cache";
import gql from "graphql-tag";

import connection from "../../lib/connection";
import { GET_TOKEN } from "../queries/local/token";
import { CreateRecordsVariables } from "../types/CreateRecords";
import { RecordsVariables } from "../types/Records";

export const typeDefs = gql`
  extend type Query {
    records(
      projectName: String!,
      streamName: String!,
      keyFields: [String!]!,
      where: JSON,
      limit: Int!,
    ): RecordsResponse!
  }

  extend type Mutation {
    createRecords(
      instanceID: UUID!,
      json: JSON!,
    ): CreateRecordsResponse!
  }

  type Record {
    recordID: ID!
    data: JSON!
    sequenceNumber: String!
  }

  type RecordsResponse {
    data: [Record!]
    error: String
  }

  type CreateRecordsResponse {
    error: String
  }
`;

export const resolvers = {
  Query: {
    records: async (_: any, { projectName, streamName, keyFields, where, limit }: RecordsVariables, { cache }: any) => {
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

      // await new Promise(function (resolve) {
      //   setTimeout(function () {
      //     resolve();
      //   }, 3000);
      // });

      // fetch
      const res = await fetch(url, { headers });
      const json = await res.json();

      // check error
      if (!res.ok) {
        return {
          __typename: "RecordsResponse",
          data: null,
          error: json.error
        };
      }

      // get data as array
      let data = json.data;
      if (!data) {
        data = [];
      } else if (!Array.isArray(data)) {
        data = [data];
      }

      // build records objects
      return {
        __typename: "RecordsResponse",
        error: null,
        data: data.map((row: any) => {
          return {
            __typename: "Record",
            recordID: makeUniqueIdentifier(keyFields, row),
            data: row,
            sequenceNumber: row["@meta"].sequence_number,
          };
        }),
      };
    },
  },
  Mutation: {
    createRecords: async (_: any, { instanceID, json }: CreateRecordsVariables, { cache }: any) => {
      // build url with limit and where
      const url = `${connection.GATEWAY_URL}/streams/instances/${instanceID}`;

      // build headers with authorization
      const headers: any = { "Content-Type": "application/json" };
      const { token } = cache.readQuery({ query: GET_TOKEN });
      if (token) {
        headers.Authorization = `Bearer ${token}`;
      }

      // submit
      const res = await fetch(url, {
        method: "POST",
        headers,
        body: JSON.stringify(json),
      });

      // check error
      let error = null;
      if (!res.ok) {
        try {
          const data = await res.json();
          error = data.error || null;
        } catch {
          error = res.text();
        }
      }

      return {
        __typename: "CreateRecordsResponse",
        error,
      };
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
