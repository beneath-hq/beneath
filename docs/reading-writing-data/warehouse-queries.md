---
title: Data warehouse queries
description: Analytical SQL queries that efficiently process entire tables
menu:
  docs:
    parent: reading-writing-data
    weight: 500
weight: 500
---

Beneath automatically writes records to a data warehouse to enable analytical SQL queries (OLAP). Check out [Unified data system]({{< ref "/docs/concepts/unified-data-system" >}}) to learn more about the different formats Beneath stores data in.

## Query examples

Assume we have a table of page views with path `example/project/page-views` and schema:

```graphql
type PageView @schema {
  user_id: Int! @key
  time: Timestamp! @key
  url: String
}
```

We can count the total number of page views:

```sql
SELECT count(*)
FROM `example/project/page-views`
```

or count the number of unique visitors per day:

```sql
SELECT timestamp_trunc(time, DAY), count(distinct user_id)
FROM `example/project/page-views`
GROUP BY timestamp_trunc(time, DAY)
```

## Query syntax

Under the hood, Beneath uses Google BigQuery as its data warehouse. It generally shares the same syntax as BigQuery, and you can use most of [BigQuery's built-in functions](https://cloud.google.com/bigquery/docs/reference/standard-sql/query-syntax).

The main deviation from BigQuery is in how you reference tables: Tables are given as paths quoted with backticks (<code>\`</code>), for example <code>\`example/project/page-views\`</code> (see examples above for more).

## Deduplicating by key

Every record in a table is added to the warehouse. That means if you write multiple records with the same schema key, they will all be processed by the query. To support deduplication of records with the same key, we include two hidden columns for each record:

- `__key` contains a bytes-serialized representation of the record's unique key
- `__timestamp` contains the write timestamp for the record

So the following pattern can be used to deduplicate records:

```sql
WITH deduplicated AS (
  SELECT r.* FROM (
    SELECT array_agg(t ORDER BY t.__timestamp DESC LIMIT 1)[OFFSET(0)] r
    FROM `username/project/table` t
    GROUP BY t.__key
  )
)
SELECT count(*)
FROM deduplicated
```
