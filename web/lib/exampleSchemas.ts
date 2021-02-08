export interface ExampleSchema {
  name: string;
  detail: string;
  schema: string;
}

const EXAMPLE_SCHEMAS = [
  {
    name: "Movies",
    language: "GraphQL",
    schema: `" Description of the stream goes here "
type Movie @schema {
  title: String! @key
  released_on: Timestamp! @key
  director: String!
  platform: Platform!
  description: String # optional field (no '!' after the type)
}

enum Platform {
  Cinema
  Apple
  Amazon
  Disney
  Netflix
}`
  },
  {
    name: "Reddit posts",
    language: "GraphQL",
    schema: `type Post @schema {
  created_on: Timestamp! @key
  id: String! @key
  author: String!
  subreddit: String!
  title: String!
  text: String
  link: String
  permalink: String
  flair: String
  is_over_18: Boolean!
  is_original_content: Boolean!
  is_self_post: Boolean!
  is_distinguished: Boolean!
  is_spoiler: Boolean!
  is_stickied: Boolean!
}`
  }
];

export default EXAMPLE_SCHEMAS;