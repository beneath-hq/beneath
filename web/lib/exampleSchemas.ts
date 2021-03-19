export interface ExampleSchema {
  name: string;
  detail: string;
  schema: string;
}

const EXAMPLE_SCHEMAS = [
  {
    name: "Movies",
    language: "GraphQL",
    schema: `" A stream of movies "
type Movie @schema {
  title: String! @key
  released_on: Timestamp! @key
  director: String
  budget_usd: Int
  rating: Float
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
  },
  {
    name: "Earthquakes",
    language: "GraphQL",
    schema: `type Earthquake @schema {
  " Time of earthquake "
  time: Timestamp! @key

  " Richter scale magnitude "
  mag: Float32

  " Location of earthquake "
  place: String!

  " Link for more information "
  detail: String
}
`
  }
];

export default EXAMPLE_SCHEMAS;