---
title: Read from a stream
description: A guide to loading, replaying or subscribing to data from an existing stream
menu:
  docs:
    parent: quick-starts
    weight: 400
weight: 400
---

This quick start will help you read data from a stream on Beneath.

## Install the Beneath SDK

If you haven't already, follow the [Install the Beneath SDK]({{< ref "/docs/quick-starts/install-sdk" >}}) quick start to install and authenticate Beneath on your local computer.

## Find a stream and head to the API tab

Browse the Beneath web [console](https://beneath.dev/?noredirect=1) and navigate to a stream you want to read from. In addition to streams in your own projects, you can browse the public streams on Beneath.

Click on the API tab where you can find auto-generated documentation.

For an example, look at the API for [earthquakes](https://beneath.dev/examples/earthquakes/stream:earthquakes/-/api) or [Reddit posts](https://beneath.dev/examples/reddit/stream:r-wallstreetbets-posts/-/api).

<video width="99%" playsinline controls>
  <source src="/media/docs/quickstart-read-stream.mp4" type="video/mp4">
</video>

## Select a code snippet

The API tab details several different ways to consume data, depending on your use case:

- Read the entire stream into memory
- Replay the stream's history and subscribe to changes
- Lookup records by key
- Analyze with SQL
