# Links and resources

Use this document to track interesting resources (articles, videos, books, etc.) on topics relevant to the development of Beneath. 

## Architecture

- (Book) [Designing Data-Intensive Applications](https://dataintensive.net/). THE book on the design, details and trade-offs of data systems at any scale.
- (Article) [Kleppmann 2019 article on OLEP paradigm](https://queue.acm.org/detail.cfm?id=3321612)

## Data serialization

- (Article) [Sort order preserving serialization](https://ananthakumaran.in/2018/08/17/order-preserving-serialization.html). We use the library described here for encoding keys in the engine.

## Analytics

- (Article) [When to track on the client vs. server](https://segment.com/academy/collecting-data/when-to-track-on-the-client-vs-server/)
- (Article) [Should I collect data on the client or server?](https://segment.com/docs/guides/best-practices/should-i-instrument-data-collection-on-the-client-or-server/)

## Data systems in action

- (Article) [Centrifuge: a reliable system for delivering billions of events per day](https://segment.com/blog/introducing-centrifuge/)
- (Article) [Delivering billions of messages exactly once](https://segment.com/blog/exactly-once-delivery/)
- (Discussion) [Ask HN: What does your BI stack look like?](https://news.ycombinator.com/item?id=21513566)

## Go

- (Example) [Standard Go Project Layout](https://github.com/golang-standards/project-layout)

## React

- (Article) [How to fetch data with React Hooks?](https://www.robinwieruch.de/react-hooks-fetch-data)

## gRPC

- (Documentation) [Guides](https://www.grpc.io/docs/guides/)
- (Documentation) [gRPC Basics - Python](https://grpc.io/docs/tutorials/basic/python/)
- (Article) [gRPC Load Balancing](https://grpc.io/blog/loadbalancing/)

## Apache Beam

- (Video) [Beam Summit Europe 2019 videos](https://www.youtube.com/playlist?list=PL4dEBWmGSIU_jJ82n0WK46agJy4ThegIQ)
- (Documentation) [Programming Guide](https://beam.apache.org/documentation/programming-guide/)
- (Code) [Python Cookbook](https://github.com/apache/beam/tree/master/sdks/python/apache_beam/examples/cookbook)
- (Documentation) [Overview: Developing a new I/O connector](https://beam.apache.org/documentation/io/developing-io-overview/)
- (Documentation) [Developing I/O connectors for Python](https://beam.apache.org/documentation/io/developing-io-python/)
- Code [Bigtable connector for Python](https://github.com/apache/beam/blob/master/sdks/python/apache_beam/io/gcp/bigtableio.py). This nicely showcases the internals of a DoFn, and implements `__getstate__` and `__setstate__`.
