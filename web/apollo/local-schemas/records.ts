import { ApolloCache, ApolloClient } from '@apollo/client';
import gql from "graphql-tag";

import { GATEWAY_URL} from "../../lib/connection";
import { GET_AID, GET_TOKEN } from "../queries/local/token";
import { CreateRecordsVariables } from "../types/CreateRecords";

export const typeDefs = gql`
  extend type Mutation {
    createRecords(
      instanceID: UUID!,
      json: JSON!,
    ): CreateRecordsResponse!
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
  Mutation: {
    createRecords: async (_: any, { instanceID, json }: CreateRecordsVariables, { cache }: ResolverContext) => {
      // build url with limit and where
      const url = `${GATEWAY_URL}/v1/-/instances/${instanceID}`;

      // build headers with authorization
      const headers: any = makeHeaders(cache);

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

const makeUniqueIdentifier = (keyFields: string[], data: any, includeTimestamp: boolean) => {
  let id = keyFields.reduce((prev, curr) => `${prev}-${data[curr]}`, "");
  if (includeTimestamp) {
    const ts = data["@meta"] && data["@meta"].timestamp;
    id = `${id}-${ts || ""}`;
  }
  return id;
};

const makeHeaders = (cache: ApolloCache<any>) => {
  const headers: any = { "Content-Type": "application/json" };
  const { aid } = cache.readQuery({ query: GET_AID }) as any;
  const { token } = cache.readQuery({ query: GET_TOKEN }) as any;
  if (aid) {
    headers["X-Beneath-Aid"] = aid;
  }
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }
  return headers;
};
