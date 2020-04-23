---
title: Unified data system
description: How and why Beneath bundles several different data technologies
menu:
  docs:
    parent: overview
    weight: 300
weight: 300
---

In order to to provide the necessary functionality, modern data systems often integrate several data technologies that serve different purposes. It quickly becomes non-trivial to integrate and maintain these technologies – indeed, that's the essence of the *data engineering* discipline. 

Beneath aims to eliminate data engineering overhead for the majority of production data science projects. So in order to provide a seamless experience, Beneath bundles three of the most common kinds of data technologies and automatically integrates them.

## Data Technologies

When you write data to Beneath, it consistently becomes available in these three systems:

- **Log-based publish/subscribe:** A system that receives and stores *messages* (also called *events*) and streams them in real-time to other systems (called *subscribers*). Examples include Apache Kafka, RabbitMQ, Google Pub/Sub and Amazon Kinesis. Also (depending on its properties) called an *event log*, *message queue* or *message broker*.
- **Data warehouse:** A system that stores data with a focus on analytical processing, for example business intelligence. It's slow if you need to quickly find individual rows, but very fast when you need to analyze an entire dataset. Examples include BigQuery, Redshift, Hive and Snowflake. Also (with some variation) called a *data lake* or *OLAP database*.
- **Operational data index:** A system that stores data with a focus on fast indexed lookups of individual records. It allows you to fetch data in milliseconds, thousand of times per second, which is useful when rendering a website or serving an API. Examples include Bigtable, Cassandra, MongoDB and Postgres. Also (depending on its properties) called a *key-value store*, *relational database* or *OLTP database*.

By automatically configuring and integrating these different technologies, Beneath allows you to get started working on your data sourcing and transformations right away. Not only do you save months of wiring data technologies, you also get a stable data system that's constantly updated to with new best practices.

## Example

To illustrate how these systems work in tandem, imagine you're building a weather forecasting website. Every time you get new weather data, you write it to Beneath and it becomes available in every system. The *publish/subscribe system* instantly pushes the data to your weather prediction model, which uses it to compute an updated forecast that it writes back into Beneath. Every time someone visits your website, you serve them the most recent forecast from the *operational data store*. Once a day, you re-train your weather prediction model with a complex query that runs in the *data warehouse*.

## Technologies Beneath Uses Under the Hood

Under the hood, the cloud version of Beneath uses Google Pub/Sub for log streaming, Google BigQuery as its data warehouse and Google Bigtable as an operational data index. If you self-host Beneath, we provide drivers for a variety of other technologies. While the choice of underlying technologies have certain implications, Beneath generally abstracts away many of the differences.

## Other Data Technologies

In addition to the three categories mentioned above, there are some rarer categories that are also worth mentioning. They include stream processing engines (for running your data transformation code), graph databases (for querying networks of data), and full-text search systems (for advanced search). Beneath doesn't currently bundle them, but we're working on changing that!
