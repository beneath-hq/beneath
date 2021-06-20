---
title: Read from a real-time table
description: A guide to replaying, streaming, filtering and analyzing data from an existing real-time table
menu:
  docs:
    parent: quick-starts
    weight: 400
weight: 400
---

This quick start will help you read data from a table on Beneath.

## Install the Beneath SDK

If you haven't already, follow the [Install the Beneath SDK]({{< ref "/docs/quick-starts/install-sdk" >}}) quick start to install and authenticate Beneath on your local computer.

## Find a table and head to the API tab

Browse the Beneath web [console](https://beneath.dev/?noredirect=1) and navigate to a table you want to read from. In addition to tables in your own projects, you can browse the public tables on Beneath.

Click on the API tab where you can find auto-generated documentation.

For an example, look at the API for [earthquakes](https://beneath.dev/examples/earthquakes/table:earthquakes/-/api) or [Reddit posts](https://beneath.dev/examples/reddit/table:r-wallstreetbets-posts/-/api).

<video width="99%" playsinline controls>
  <source src="/media/docs/quickstart-read-table.mp4" type="video/mp4">
</video>

## Select a code snippet

The API tab details several different ways to consume data, depending on your use case:

- **Read the entire table into memory**
- **Replay the table's history and subscribe to changes**
- **Lookup records by key**
- **Analyze with SQL**
