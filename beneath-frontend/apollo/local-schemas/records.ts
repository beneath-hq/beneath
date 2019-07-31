import gql from "graphql-tag";

export const typeDefs = gql`
  extend type Query {
    records(instanceID: UUID!): [Record!]!
  }

  type Record {
    recordID: ID!
    data: String!
    sequenceNumber: String!
  }
`;

export const resolvers = {
  Query: {
    records: async (_: any, { instanceID }: any, { cache }: any) => {
      // const { cartItems } = cache.readQuery({ query: GET_CART_ITEMS });
      // const data = {
      //   cartItems: cartItems.includes(id)
      //     ? cartItems.filter(i => i !== id)
      //     : [...cartItems, id],
      // };
      // cache.writeQuery({ query: GET_CART_ITEMS, data });
      // return data.cartItems;

      return [{
        __typename: "Record",
        recordID: "aaaa",
        data: "fjskdpamf",
        sequenceNumber: "fslak",
      }];
    },
  },
};

export default {
  typeDefs,
  resolvers,
};
