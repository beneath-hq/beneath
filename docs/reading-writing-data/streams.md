---
title: Streams
description: An explanation of how streams work in Beneath
menu:
  docs:
    parent: billing
    weight: 100
weight: 100
---

To clear out any confusion: Beneath has a broad interpretation of the term "stream" that includes *bounded* data collections (also known as *batch data* or *static data*). This interpretation is based on the philosophy that even bounded collections comes into existance, and will in turn be consumed, in some kind of sequence. 

All `stream`s have an associated schema that defines the fields of its records. The schema must also define a *key* comprised of one or more columns (also known as a *primary key* or *unique key*). Beneath rejects records that do not adhere to the `stream`'s schema.

Records are not stored directly in a `stream`, but rather in a `stream instance`, which a `stream` can have many of. You can think of them as different versions of the `stream`. A `stream instance` can be marked as *finalized*, which indicates that no more records will be written to it. The following list gives some examples that clarify the usefulness of `stream instance`s:

- A stream that represents *unbounded data* (such as trades on a stock exchange) will only have one *unfinalized* `stream instance` that grows forever.
- A stream that represents *batch-updated data* (such as the results of an SQL query that runs once every day) will have a new `stream instance` for every batch. When a new batch has been completely written, the `stream instance` is marked as *finalized*. Old `stream instance`s for previous batches are deleted.
- A stream that represents *static data* (such as reference data that never changes) will have one *finalized* `stream instance` that never changes.

Every record in a `stream instance` has an associated timestamp. When you write a record, you can specify a custom timestamp or use the current timestamp of the Beneath server. The timestamp comes in handy when when several records with the same *key* are written to a `stream instance`. In these cases, the record with the highest timestamp is indexed in the operational data cache.

TODO:

## Root streams vs. derived streams

<em>Root</em> streams are those data streams that are directly written to Beneath from an external source. 
Additionally, root streams can be <em>manual</em>. Manual streams permit the user to upload the data in batch via the Data Terminal's UI. Streams that are not manual accept data programatically.

<em>Derived</em> streams are those data streams that are the result of stream processing. Applying a data transformation (a "model") to a Beneath stream results in a new, derived stream.

