---
title: Use the Pipeline API
description: A guide to generating and processing streams with the Pipeline API
menu:
  docs:
    parent: quick-starts
    weight: 500
weight: 500
---

Beneath Pipelines provide an abstraction over the basic Beneath APIs that makes it easier to develop, test, and deploy stream processing logic.

Beneath pipelines are currently quite basic and do not yet support joins and aggregations. They are still well-suited for generating streams, one-to-N stream derivation, as well as syncing and alerting records.

## Install the Beneath SDK

If you haven't already, follow the [Install the Beneath SDK]({{< ref "/docs/quick-starts/install-sdk" >}}) quick start to install and authenticate Beneath on your local computer.

## Find a stream and head to API "Pipelines" tab

Browse the Beneath web [console](https://beneath.dev/?noredirect=1) and navigate to a stream you want to process.

Click through to the stream's "Pipelines" API: API > Python > Pipelines. There you'll find a stream-specific tutorial on how to create a pipeline.

For an example, look at the API for [earthquakes](https://beneath.dev/examples/earthquakes/stream:earthquakes/-/api?language=python&action=pipelines) or [r/wallstreetbets posts](https://beneath.dev/examples/reddit/stream:r-wallstreetbets-posts/-/api?language=python&action=pipelines) streams.

<video width="99%" playsinline controls>
  <source src="/media/docs/quickstart-create-pipeline.mp4" type="video/mp4">
</video>

## Select a code template and insert your logic

The Pipelines tab has two code templates:

- **Generate records for this stream**. Follow this guide to consume an external data source and write the data to Beneath. You can also adapt the `stream_path` and `schema` params to make the pipeline produce an entirely new stream.
- **Derive a new stream**. Follow this guide to apply processing logic, such as filtering or enrichment, to the stream.

## Run and deploy your pipeline

At the bottom of the Pipelines tab, you'll find instructions for how to run and deploy a pipeline. Once your pipeline is live, use the web console to view the pipeline's output and monitor its activity.
