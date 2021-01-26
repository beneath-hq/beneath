---
title: Links and resources
description:
menu:
  docs:
    parent: contributing
    weight: 900
weight: 900
---

Use this document to track interesting resources (articles, videos, books, etc.) on topics relevant to the development of Beneath.

## Architecture

- (Book) [Designing Data-Intensive Applications](https://dataintensive.net/). THE book on the design, details and trade-offs of data systems at any scale.
- (Article) [Kleppmann 2019 article on OLEP paradigm](https://queue.acm.org/detail.cfm?id=3321612)

## Workflow

- (Article) [Design Docs at Google](https://www.industrialempathy.com/posts/design-docs-at-google/)

## Data serialization

- (Article) [Sort order preserving serialization](https://ananthakumaran.in/2018/08/17/order-preserving-serialization.html). We use the library described here for encoding keys in the engine.

## Analytics

- (Article) [When to track on the client vs. server](https://segment.com/academy/collecting-data/when-to-track-on-the-client-vs-server/)
- (Article) [Should I collect data on the client or server?](https://segment.com/docs/guides/best-practices/should-i-instrument-data-collection-on-the-client-or-server/)

## Data systems in action

- (Article) [Centrifuge: a reliable system for delivering billions of events per day](https://segment.com/blog/introducing-centrifuge/)
- (Article) [Delivering billions of messages exactly once](https://segment.com/blog/exactly-once-delivery/)
- (Paper) [Large-scale Incremental Processing Using Distributed Transactions and Notifications](https://storage.googleapis.com/pub-tools-public-publication-data/pdf/36726.pdf)
- (Discussion) [Ask HN: What does your BI stack look like?](https://news.ycombinator.com/item?id=21513566)

## Go

- (Example) [Standard Go Project Layout](https://github.com/golang-standards/project-layout)

## React

- (Article) [How to fetch data with React Hooks?](https://www.robinwieruch.de/react-hooks-fetch-data)

## gRPC

- (Documentation) [Guides](https://www.grpc.io/docs/guides/)
- (Documentation) [gRPC Basics - Python](https://grpc.io/docs/tutorials/basic/python/)
- (Article) [gRPC Load Balancing](https://grpc.io/blog/loadbalancing/)

## Log streaming

- (Article) [The Log: What every software engineer should know about real-time data's unifying abstraction](https://engineering.linkedin.com/distributed-systems/log-what-every-software-engineer-should-know-about-real-time-datas-unifying)
- (Article) [Online Event Processing](https://queue.acm.org/detail.cfm?id=3321612)
- (Article) [Kafka Exactly Once Semantics](https://hevodata.com/blog/kafka-exactly-once/). Decent introduction to idempotency and atomic transactions in Kafka.
- (Documentation) [Docs for `aiokafka`](https://aiokafka.readthedocs.io/en/stable/index.html). Good Python examples of stream processing with `await`s in Python

## Apache Beam

- (Video) [Beam Summit Europe 2019 videos](https://www.youtube.com/playlist?list=PL4dEBWmGSIU_jJ82n0WK46agJy4ThegIQ)
- (Documentation) [Programming Guide](https://beam.apache.org/documentation/programming-guide/)
- (Code) [Python Cookbook](https://github.com/apache/beam/tree/master/sdks/python/apache_beam/examples/cookbook)
- (Documentation) [Overview: Developing a new I/O connector](https://beam.apache.org/documentation/io/developing-io-overview/)
- (Documentation) [Developing I/O connectors for Python](https://beam.apache.org/documentation/io/developing-io-python/)
- (Code) [Bigtable connector for Python](https://github.com/apache/beam/blob/master/sdks/python/apache_beam/io/gcp/bigtableio.py). This nicely showcases the internals of a DoFn, and implements `__getstate__` and `__setstate__`.
- (Article) [Apache Beam: How Beam Runs on Top of Flink](https://flink.apache.org/ecosystem/2020/02/22/apache-beam-how-beam-runs-on-top-of-flink.html). Explains how Beam pipelines written in different languages all can run on Flink.
