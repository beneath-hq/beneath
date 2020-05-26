---
title: Creating streams
description: How to define schemas and create streams
menu:
  docs:
    parent: reading-writing-data
    weight: 200
weight: 200
---

## Create a stream in Python or with the CLI

1. If you haven't already, [install and setup the Beneath CLI]({{< ref "/docs/managing-resources/cli.md" >}}).
2. If you haven't already, create a project using the command-line:
```bash
beneath project create USERNAME/NEW_PROJECT_NAME
```
3. Define a schema for your stream (see the next section for details)
4. You now have two options for creating a stream:
    1. Using the command line. Save your schema to a file (e.g. `schema.gql`) and run:
    ```bash
    beneath stream stage USERNAME/PROJECT/NEW_STREAM_NAME -f schema.gql
    ```
    2. Using Python. Copy and paste the following snippet:
    ```python
    import beneath
    client = beneath.Client()
    stream = await client.stage_stream('USERNAME/PROJECT/NEW_STREAM_NAME', """
      type Example @stream @key(fields: ["foo"]) {
        foo: Int!
        bar: String!
        foo_bar: Float
      }
    """)
    ```

To delete a stream, run the following command:
```bash
beneath stream delete USERNAME/PROJECT/STREAM
```

To learn more about using the Python library, consult the ([reference documentation](https://python.docs.beneath.dev/)).

## Defining stream schemas

Every stream in Beneath has a schema defined with a variant of the [GraphQL schema definition language](https://graphql.org/learn/schema/). Here is an example:
```graphql
type Example @stream @key(fields: ["foo"]) {
  foo: Int!
  bar: String!
  foo_bar: Float
}
```
Beneath enforces some special conventions on top of the normal GraphQL language:
- Field names must be specified in snake_case and type names in PascalCase
- The stream schema should have an `@stream` annotation and an `@key(fields: [...])` annotation to indicate the field(s) that make up the stream's unique key
- Fields in the unique key must be marked required with an exclamation mark (e.g. `foo: Int!`)
- The supported primitive types are:

| Type | Definition |
|---|---|
| `Boolean` | True or false |
| `Int64` or `Int` | A 64-bit whole number |
| `Int32` | A 32-bit whole number |
| `Float64` or `Float` | A 64-bit (double precision) IEEE 754 floating-point number |
| `Float32` | A 32-bit (single precision) IEEE 754 floating-point number |
| `String` | A variable-length sequence of unicode characters |
| `Bytes` | A variable-length sequence of bytes |
| `BytesN` | A fixed-length sequence of *N* bytes, e.g. `Bytes32` |
| `Numeric` | An arbitrarily large whole number |
| `Timestamp` | A millisecond-precision UTC date and time (no time zone) |

- You can define custom sub-types and enums, for example:
```graphql
type Place @stream @key(fields: ["place_id"]) {
  place_id: Int!
  location: Point!
  kind: PlaceKind!
}

type Point {
  x: Float!
  y: Float!
}

enum PlaceKind {
  Cafe
  Restaurant
  IceCreamVendor
}
```
