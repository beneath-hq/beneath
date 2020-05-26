---
title: Understanding streams
description: Everything you need to understand about streams in Beneath
menu:
  docs:
    parent: reading-writing-data
    weight: 100
weight: 100
---

To clear out any confusion: Beneath has a broad interpretation of the term "stream" that includes *bounded* data collections (also known as *batch data* or *static data*). This interpretation is based on the philosophy that even bounded collections comes into existance, and will in turn be consumed, in some kind of sequence.

Streams in Beneath have some unique properties that you can learn more about below.

## Streams have a schema and a unique key

All *streams* have an associated schema that defines the fields of its records. The schema must also define a *key* comprised of one or more columns (also known as a *primary key* or *unique key*). Beneath rejects records that do not adhere to the stream's schema. 

## Streams can have multiple versions, known as *instances*

Records are not stored directly in a *stream*, but rather in a *stream instance*, which a stream can have many of. You can think of them as different versions of the stream.

Every stream instance has two boolean flags: a) It can be marked as *finalized*, which indicates that no more records will be written to it, and b) it can be marked as *primary*, which makes it the default instance for the stream, causing it to be shown in the Terminal.

The following list gives some examples that clarify the usefulness of *stream instances*:

- A stream that represents *unbounded data* (such as trades on a stock exchange) will have one *unfinalized* stream instance that grows forever.
- A stream that represents *batch-updated data* (such as the results of an SQL query that runs once every day) will have a new stream instance for every batch. When a new batch has been completely written, the stream instance is marked as *finalized*. Old stream instances for previous batches are deleted.
- A stream that represents *static data* (such as reference data that never changes) will have one *finalized* stream instance that never changes.

## All records have an associated timestamp

Every record in a *stream instance* has an associated timestamp. When you write a record, you can specify a custom timestamp or use the current timestamp of the Beneath server (the default). The timestamp comes in handy when several records with the same *key* are written to a stream instance. In these cases, the record with the highest timestamp is added in the operational data index.

## Streams are replicated in a streaming log, an operational data index and a data warehouse

Records written to Beneath are automatically made stored in several systems, which are useful for different purposes. These are:

- A persistant and streaming log, for replaying the history of a stream instance and staying subscribed to new changes
- An operational data index, which stores data for low-latency (milliseconds) filtered lookups based on the unique key 
- A data warehouse, for analytical SQL-based queries that process a large slice of the stream instance in one go

To learn more about these systems, see [unified data system]({{< ref "/docs/introduction/unified-data-system" >}}).

The following pages in this section explain more about how to use these systems in practice.
