import { ApolloClient } from "apollo-boost";
import { ApolloCache } from "apollo-cache";
import gql from "graphql-tag";

import connection from "../../lib/connection";
import { GET_TOKEN } from "../queries/local/token";
import { QUERY_STREAM } from "../queries/stream";
import { CreateRecordsVariables } from "../types/CreateRecords";
import { LatestRecordsVariables } from "../types/LatestRecords";
import { RecordsVariables } from "../types/Records";

export const typeDefs = gql`
  extend type Query {
    records(
      projectName: String!,
      streamName: String!,
      where: JSON,
      after: JSON,
      limit: Int!,
    ): RecordsResponse!

    latestRecords(
      projectName: String!,
      streamName: String!,
      limit: Int!,
    ): [Record!]!
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

interface ResolverContext {
  client: ApolloClient<any>;
  cache: ApolloCache<any>;
}

export const resolvers = {
  Query: {
    records: async (_: any, args: RecordsVariables, { cache, client }: ResolverContext) => {
      // get stream
      const { data: { stream } } = await  client.query({
        query: QUERY_STREAM,
        variables: { projectName: args.projectName, name: args.streamName },
      });
      if (!stream) {
        return {
          __typename: "RecordsResponse",
          data: null,
          error: "couldn't find stream",
        };
      }

      // build url with limit and where
      let url = `${connection.GATEWAY_URL}/projects/${args.projectName}/streams/${args.streamName}`;
      url += `?limit=${args.limit}`;
      if (args.where) {
        url += `&where=${JSON.stringify(args.where)}`;
      }
      if (args.after) {
        url += `&after=${JSON.stringify(args.after)}`;
      }

      // build headers with authorization
      const headers: any = { "Content-Type": "application/json" };
      const { token } = cache.readQuery({ query: GET_TOKEN }) as any;
      if (token) {
        headers.Authorization = `Bearer ${token}`;
      }

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
            recordID: makeUniqueIdentifier(stream.keyFields, row),
            data: row,
            sequenceNumber: row["@meta"].sequence_number,
          };
        }),
      };
    },
    latestRecords: async (_: any, args: LatestRecordsVariables, { cache, client }: ResolverContext) => {
      // TODO:
      const res = await resolvers.Query.records(null, args, { cache, client });
      return res.data;
    },
  },
  Mutation: {
    createRecords: async (_: any, { instanceID, json }: CreateRecordsVariables, { cache }: ResolverContext) => {
      // build url with limit and where
      const url = `${connection.GATEWAY_URL}/streams/instances/${instanceID}`;

      // build headers with authorization
      const headers: any = { "Content-Type": "application/json" };
      const { token } = cache.readQuery({ query: GET_TOKEN }) as any;
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
  return keyFields.reduce((prev, curr) => `${prev}-${data[curr]}`, "");
};
