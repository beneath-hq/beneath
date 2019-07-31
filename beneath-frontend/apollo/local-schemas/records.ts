import gql from "graphql-tag";

import connection from "../../lib/connection";

export const typeDefs = gql`
  extend type Query {
    records(
      projectName: String!,
      streamName: String!,
      keyFields: [String!]!
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
    records: async (_: any, { projectName, streamName, keyFields }: any, { cache }: any) => {
      const url = `${connection.GATEWAY_URL}/projects/${projectName}/streams/${streamName}`;
      const res = await fetch(url);
      const json = await res.json();
      const records = json.map((row: any) => {
        return {
          __typename: "Record",
          recordID: makeUniqueIdentifier(keyFields, row),
          data: row,
          sequenceNumber: row["@meta"].sequence_number,
        };
      });

      // const { cartItems } = cache.readQuery({ query: GET_CART_ITEMS });
      // const data = {
      //   cartItems: cartItems.includes(id)
      //     ? cartItems.filter(i => i !== id)
      //     : [...cartItems, id],
      // };
      // cache.writeQuery({ query: GET_CART_ITEMS, data });
      // return data.cartItems;

      return records;
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
