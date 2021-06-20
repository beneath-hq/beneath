---
title: Schemas and keys
description: How to define schemas and keys
menu:
  docs:
    parent: reading-writing-data
    weight: 200
weight: 200
---

If you're looking for an introduction to tables in Beneath, head over to the [Concepts]({{< ref "/docs/concepts" >}}) section.

## Schemas

Every table in Beneath has a _schema_ and a _key_:

- The **schema** defines fields and data types for records in the table. It is a variant of the [GraphQL](https://graphql.org/learn/schema/) schema definition language.
- The **key** defines one or more fields that uniquely identify the record (also known as a _primary key_ or _unique key_). It is used for log compaction and table indexing.

Here is an example:

```graphql
type Click @schema {
  user_id: Int! @key
  time: Timestamp! @key
  label: String!
  details: String
}
```

Beneath enforces some special conventions on top of the normal GraphQL language:

- Field names must be specified in `snake_case` and type names in `PascalCase`
- The schema type should have an `@schema` annotation
- The `@key` annotation specifies the key field(s), where multiple `@key` annotations create a composite key in the respective order
- Key fields must be marked required with an exclamation mark (e.g. `foo: Int!`)

## Primitive types

The supported primitive types are:

| Type                 | Definition                                                 |
| -------------------- | ---------------------------------------------------------- |
| `Boolean`            | True or false                                              |
| `Int` or `Int64`     | A 64-bit whole number                                      |
| `Int32`              | A 32-bit whole number                                      |
| `Float` or `Float64` | A 64-bit (double precision) IEEE 754 floating-point number |
| `Float32`            | A 32-bit (single precision) IEEE 754 floating-point number |
| `String`             | A variable-length sequence of unicode characters           |
| `Bytes`              | A variable-length sequence of bytes                        |
| `BytesN`             | A fixed-length sequence of _N_ bytes, e.g. `Bytes32`       |
| `Numeric`            | An arbitrarily large whole number                          |
| `Timestamp`          | A millisecond-precision UTC date and time                  |

## Nested types

You can define nested types (the `@schema` annotation indicates the main type):

```graphql
type Place @schema {
  place_id: Int! @key
  location: Point!
  label: String
}

type Point {
  x: Float!
  y: Float!
}
```

However, we generally encourage flat schemas as they are easier to analyze with SQL and look nicer in tables (e.g., in the example, `location` could be flattened to `location_x` and `location_y`).

## Enums

You can define enums, for example:

```graphql
type Journey @schema {
  journey_id: String! @key
  kind: Vehicle!
  distance_km: Int
}

enum Vehicle {
  Bike
  Bus
  Car
  Plane
  Train
}
```
