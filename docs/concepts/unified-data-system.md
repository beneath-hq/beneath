---
title: Unified data system
description: How and why Beneath bundles several different data technologies
menu:
  docs:
    parent: concepts
    weight: 100
weight: 100
---

We have found that many production data science projects need to integrate several different data technologies to adequately consume and query data. To provide a seamless experience, when you write data to a stream in Beneath, it automatically replicates it to:

- a **streaming log** for replay/subscribe (e.g. to sync or enrich data)
- a **data warehouse** for OLAP queries with SQL (e.g. to render dashboards)
- an **operational data store** for scalable, indexed lookups (e.g. to render data in your frontend)

Beneath lets you access these sources through a single layer of abstraction, so you can focus on your use cases without getting bogged down in integration and maintenance. The following section explains each of these technologies in more detail.

## Streaming log

The streaming log keeps real-time, ordered track of every record written to a stream, allowing you to read new data starting from any point in time _without missing a single change_.

In practice, it means you can replay the history of a stream as if you had been subscribed since its beginning, then stay subscribed to get every new change within milliseconds of it happening. If your code is down for a while or only runs periodically, you can get every change that happened in the meantime once you resubscribe (it's an _at-least-once guarantee_).

(Systems that can serve as a streaming log are sometimes called an _event log_ or _message queue_, and stand-alone implementations include Apache Kafka, Amazon Kinesis, Cloud Pubsub and RabbitMQ).

## Data warehouse

The data warehouse stores records with a focus on analytical processing with SQL, making it ideal for business intelligence and ad-hoc exploration. It's slow for finding individual records, but lets you scan and analyze an entire stream in seconds.

(Systems that serve as a data warehouse are sometimes called a _data lake_ or _OLAP database_, and stand-alone implementations include BigQuery, Snowflake, Redshift and Hive).

## Operational data store

The operational data store enables fast, indexed lookups of individual records or specific ranges of records. It allows you to fetch records in milliseconds, thousand of times per second, which is useful when rendering a website or serving an API.

In Beneath, records are currently indexed based on their unique key (see [Streams]({{< ref "/docs/concepts/streams" >}}) for more). For streams that contain multiple records with the same unique key (for example due to updates), the operational data store only indexes the most recent record.

(While broader categories, _key-value stores_ and _OLTP databases_ often serve as operational data stores, and popular stand-alone implementations include MongoDB, Postgres, Cassandra and Bigtable).

## Example

To illustrate how these systems work in tandem, imagine you're building a weather forecasting website. Every time you get new weather data, you write it to Beneath and it becomes available in every system. The _streaming log_ instantly pushes the data to your weather prediction model, which uses it to compute an updated forecast that it writes back into Beneath. Every time someone visits your website, you serve them the most recent forecast from the _operational data index_. Once a day, you re-train your weather prediction model with a complex SQL query that runs in the _data warehouse_.

## Technologies Beneath uses under the hood

We're not interested in reinventing the wheel, so under the hood, Beneath uses battle-tested data technologies. The cloud version of Beneath uses a combination of Google Bigtable and Cloud Pub/Sub for log streaming, Google BigQuery as its data warehouse and Google Bigtable as an operational data store. If you self-host Beneath, we provide drivers for a variety of other technologies. While the choice of underlying technologies have certain implications, Beneath generally abstracts away many of the differences.

## Other data technologies

In addition to the three data technologies mentioned above, there are some rarer technologies worth mentioning, such as graph databases (for querying networks of data) and full-text search systems (for advanced search). We're devoted to covering more data access paradigms, so if Beneath doesn't currently serve your use case, we would love to [hear from you]({{< ref "/contact" >}}).
