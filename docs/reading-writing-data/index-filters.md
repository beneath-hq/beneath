---
title: Filtered index lookups
description: How to quickly retrieve indexed records from tables
menu:
  docs:
    parent: reading-writing-data
    weight: 400
weight: 400
---

Beneath automatically indexes records for fast lookup and filtering by key. Check out [Unified data system]({{< ref "/docs/concepts/unified-data-system" >}}) to learn more about the different formats Beneath stores data in.

## What can you filter?

Indexes in Beneath enable fast (milliseconds) single-key lookup and range-based filtering. The index only contains one record for each key -- if a table contains multiple records with the same key, the index will contain the most recent record for the key.

You can only filter records based on their schema key (see [Schema definition]({{< ref "/docs/reading-writing-data/schema-definition" >}})).

For schema keys composed of multiple fields, the index must be used left-to-right. For example, if the schema key is `(user_id, time)`, you can only filter by time for a specific user, not all users.

## Filter syntax

We use a JSON-based query syntax for filters. Assume we have a schema:

```graphql
type StockPrice @schema {
  symbol: String! @key
  time: Timestamp! @key
  price_usd: Float
}
```

Filtering for a key field:

```json
{ "symbol": "AAPL" }
```

returns all stock prices for symbol `AAPL`.

Filtering for a compound key:

```json
{ "symbol": "AAPL", "day": "2020-03-11T19:00:00" }
```

returns the stock price for symbol `AAPL` at 7pm on 11th March 2020.

Filtering for a range in a compound key:

```json
{ "symbol": "AAPL", "day": { "_gte": "2020-03-09", "_lt": "2020-03-16" } }
```

returns all stock prices for symbol `AAPL` between 9th and 16th March 2020 (`_gte` means "greater than or equal to" and `_lt` means "less than").

Filtering by prefix:

```json
{ "symbol": { "_prefix": "A" } }
```

returns all stock prices for all symbols starting with `A`.

## Representing data types as JSON

Not every schema field type can be natively represented in JSON. The table below shows how to encode Beneath's supported data types in JSON:

| Type        | Definition                                                                                                                                                                                                                                                                 |
| ----------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `Boolean`   | Supported in JSON                                                                                                                                                                                                                                                          |
| `Int`       | Supported in JSON                                                                                                                                                                                                                                                          |
| `Float`     | Supported in JSON                                                                                                                                                                                                                                                          |
| `String`    | Supported in JSON                                                                                                                                                                                                                                                          |
| `Bytes`     | Base64-encoded string                                                                                                                                                                                                                                                      |
| `Numeric`   | String containing an integer                                                                                                                                                                                                                                               |
| `Timestamp` | Multiple options:<ul><li>Integer of milliseconds since 1970 (Unix time)</li><li>String with date, time and timezone: `"2006-01-02T15:04:05Z07:00"` (RFC3339)</li><li>String with date and time: `"2006-01-02T15:04:05"`</li><li>String with date: `"2006-01-02"`</li></ul> |
| Enums       | String containing the enum constant                                                                                                                                                                                                                                        |
