# `engine/driver/`

Drivers connect Beneath functionality to the underlying features offered by a data management system. They allow Beneath to offer a uniform user experience while benefiting from the unique advantages of different data systems.

## Driver types

Fundamentally, Beneath currently distinguishes between three types of infrastructure components:

- `MessageQueue`: a message bus that supports at-least-once delivery
- `LookupService`: a low-latency read/write service for storing log data and indexes; must support range reads
- `WarehouseService`: an OLAP service that can execute large (long-running) analytical queries efficiently

The features expected of each of these types are described in more detail in the driver interfaces in `engine/driver/driver.go`.

Note that one underlying data system may implement several types at once. For example, at smaller scales of data (100s GB), a Postgres database could efficiently serve as both a message queue, lookup service and warehouse service.

In the future, we may want to extend this list with fundamentally different data paradigms, such as a full-text search service or a graph data service.

## Existing drivers

- `bigquery` implements Google Cloud BigQuery as a highly-scalable `WarehouseService` driver.
- `bigtable` implements Google Cloud Bigtable as a highly-scalable `LookupService` driver.
- `mock` implements a `WarehouseService` driver that simply discards data and pretends to have successfully stored it.
- `pubsub` implements Google Cloud Pub/Sub as a highly-scalable `MessageQueue` driver.
- `postgres` has not yet been implemented, but is intended to implement all driver types on Postgres for use in development and small-scale (100s GB) self-hosting deployments.

## Implementing a new driver

To implement a new driver, add a package named after the system to `engine/driver/`, and implement the relevant driver interfaces in `engine/driver/driver.go`. Look at existing drivers for inspiration.

Here's a list of drivers we're very interested in developing

- Postgres as a driver for `MessageQueue`, `LookupService` and `WarehouseService` (as described above)
- Apache Cassandra as a `LookupService`
- Apache Kafka as a `MessageQueue` and possibly a `LookupService` (leveraging Kafka Tables and/or ksqlDB)
- Amazon Redshift as a `WarehouseService`
