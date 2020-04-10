---
title: Streams
description:
menu:
  docs:
    parent: core-resources
    weight: 130
weight: 130
---
Streams are data streams that run on Beneath. Whether a data set is an event stream or batch upload, Beneath interprets all data sets as "streams." 
Even a batch upload has an associated timestamp, and future uploads or revisions of that data set will also have associated timestamps.

##### Schema
An important feature of Beneath is that every stream requires a pre-defined schema. Required schema ensures that streams only accept data that is expected (i.e. data that conforms to the schema), and all other non-conforming data is rejected. In order to create any stream, you must provide a schema .graphql file. Read more about the Beneath schema language <a href="/docs/schema-language">here</a>. 

##### Root streams vs. derived streams
<em>Root</em> streams are those data streams that are directly written to Beneath from an external source. 
Additionally, root streams can be <em>manual</em>. Manual streams permit the user to upload the data in batch via the Data Terminal's UI. Streams that are not manual accept data programatically.

<em>Derived</em> streams are those data streams that are the result of stream processing. Applying a data transformation (a "model") to a Beneath stream results in a new, derived stream.

##### CLI commands
With the CLI, you can do the following:

- list all streams in a given project
- create a root stream
- update the schema of a root stream

See the manual:
```bash
beneath stream -h
beneath root-stream -h
```
