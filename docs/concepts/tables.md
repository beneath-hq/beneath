---
title: Real-time tables and table versions
description: Everything you need to understand about tables in Beneath
menu:
  docs:
    parent: concepts
    weight: 200
weight: 200
---

Real-time tables in Beneath have some unique properties that you can learn more about below.

## Streams are replicated in a streaming log, a data warehouse and an operational data store

Records written to Beneath are automatically stored in a streaming log, a data warehouse and an operational data store. That means you can consume and query tables in the way that makes the most sense for your use case, including streaming changes, running OLAP queries, and running fast filter queries. You can learn a lot more about this feature in [Unified data system]({{< ref "/docs/concepts/unified-data-system" >}}).

## Tables have a schema and a unique key

All _tables_ have an associated schema that defines the fields of its records. The schema must also define a _key_ comprised of one or more columns (also known as a _primary key_ or _unique key_). Beneath rejects records that do not adhere to the table's schema. See [Schema definition]({{< ref "/docs/reading-writing-data/schema-definition" >}}) for more details.

## Tables can have multiple versions, known as _instances_

Records are not stored directly in a _table_, but rather in a _table instance_, which a table can have many of. Each instance has an associated _version_ number, which is a positive integer.

Every table instance has two boolean flags: a) It can be marked as _finalized_, which indicates that no more records will be written to it, and b) it can be marked as _primary_, which makes it the default instance for the table.

The following list gives some examples that clarify the usefulness of _table instances_:

- A table that represents _unbounded data_ (such as trades on a stock exchange) will have one _unfinalized_ table instance that grows forever.
- A table that represents _batch-updated data_ (such as the results of an SQL query that runs once every day) will have a new table instance for every batch. When a new batch has been completely written, the table instance is marked as _finalized_. Old table instances for previous batches are deleted.
- A table that represents _static data_ (such as reference data that never changes) will have one _finalized_ table instance that never changes.

## All records have an associated timestamp

Every record in a _table instance_ has an associated timestamp. When you write a record, you can specify a custom timestamp or use the current timestamp of the Beneath server (the default). The timestamp comes in handy when several records with the same _key_ are written to a table instance. In these cases, the record with the highest timestamp is added in the operational data index.
