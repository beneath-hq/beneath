---
title: Use the Pipeline API
description: A guide to generating and processing real-time tables with the Pipeline API
menu:
  docs:
    parent: quick-starts
    weight: 500
weight: 500
---

Beneath Pipelines provide an abstraction over the basic Beneath APIs that makes it easier to develop, test, and deploy stream processing logic.

Beneath pipelines are currently quite basic and do not yet support joins and aggregations. They are still well-suited for generating tables, one-to-N table derivation, as well as syncing and alerting records.

## Install the Beneath SDK

If you haven't already, follow the [Install the Beneath SDK]({{< ref "/docs/quick-starts/install-sdk" >}}) quick start to install and authenticate Beneath on your local computer.

## Find a table and head to API "Pipelines" tab

Browse the Beneath web [console](https://beneath.dev/?noredirect=1) and navigate to a table you want to process.

Click through to the table's "Pipelines" API: API > Python > Pipelines. There you'll find a table-specific tutorial on how to create a pipeline.

For an example, look at the API for [earthquakes](https://beneath.dev/examples/earthquakes/table:earthquakes/-/api?language=python&action=pipelines) or [r/wallstreetbets posts](https://beneath.dev/examples/reddit/table:r-wallstreetbets-posts/-/api?language=python&action=pipelines) tables.

<video width="99%" playsinline controls>
  <source src="/media/docs/quickstart-create-pipeline.mp4" type="video/mp4">
</video>

## Select a code template and insert your logic

The Pipelines tab has two code templates:

- **Generate records for this table**. Follow this guide to consume an external data source and write the data to Beneath. You can also adapt the `table_path` and `schema` params to make the pipeline produce an entirely new table.
- **Derive a new table**. Follow this guide to apply processing logic, such as filtering or enrichment, to the table.

## Run and deploy your pipeline

At the bottom of the Pipelines tab, you'll find instructions for how to run and deploy a pipeline. Once your pipeline is live, use the web console to view the pipeline's output and monitor its activity.
