---
title: Streams and stream instances
description: Everything you need to understand about streams in Beneath
menu:
  docs:
    parent: concepts
    weight: 200
weight: 200
---

To clear out any confusion: Beneath has a broad interpretation of the term "stream" that includes _bounded_ data collections (also known as _batch data_ or _static data_). This interpretation is based on the philosophy that even bounded collections comes into existance, and will in turn be consumed, in some kind of sequence.

Streams in Beneath have some unique properties that you can learn more about below.

## Streams are replicated in a streaming log, a data warehouse and an operational data store

Records written to Beneath are automatically stored in a streaming log, a data warehouse and an operational data store. This means you can consume and query streams in the way that makes the most sense for your use case. You can learn a lot more about this feature in [Unified data system]({{< ref "/docs/concepts/unified-data-system" >}}).

## Streams have a schema and a unique key

All _streams_ have an associated schema that defines the fields of its records. The schema must also define a _key_ comprised of one or more columns (also known as a _primary key_ or _unique key_). Beneath rejects records that do not adhere to the stream's schema. See [Schema definition]({{< ref "/docs/reading-writing-data/schema-definition" >}}) for more details.

## Streams can have multiple versions, known as _instances_

Records are not stored directly in a _stream_, but rather in a _stream instance_, which a stream can have many of. Each instance has an associated _version_ number, which is a positive integer.

Every stream instance has two boolean flags: a) It can be marked as _finalized_, which indicates that no more records will be written to it, and b) it can be marked as _primary_, which makes it the default instance for the stream.

The following list gives some examples that clarify the usefulness of _stream instances_:

- A stream that represents _unbounded data_ (such as trades on a stock exchange) will have one _unfinalized_ stream instance that grows forever.
- A stream that represents _batch-updated data_ (such as the results of an SQL query that runs once every day) will have a new stream instance for every batch. When a new batch has been completely written, the stream instance is marked as _finalized_. Old stream instances for previous batches are deleted.
- A stream that represents _static data_ (such as reference data that never changes) will have one _finalized_ stream instance that never changes.

## All records have an associated timestamp

Every record in a _stream instance_ has an associated timestamp. When you write a record, you can specify a custom timestamp or use the current timestamp of the Beneath server (the default). The timestamp comes in handy when several records with the same _key_ are written to a stream instance. In these cases, the record with the highest timestamp is added in the operational data index.
