<p align="center">
  <a href="https://about.beneath.dev/?utm_source=github&utm_medium=logo" target="_blank">
    <img src="https://github.com/beneath-hq/beneath/tree/master/assets/logo/banner-icon-text-background.png" alt="Beneath" height="200">
  </a>
</p>

<hr />

[![Go Report Card](https://goreportcard.com/badge/gitlab.com/beneath-hq/beneath?style=flat-square)](https://goreportcard.com/report/gitlab.com/beneath-hq/beneath)
[![GoDoc](https://godoc.org/gitlab.com/beneath-hq/beneath?status.svg)](https://godoc.org/gitlab.com/beneath-hq/beneath)
[![Twitter](https://img.shields.io/badge/Follow-BeneathHQ-blue.svg?style=flat&logo=twitter)](https://twitter.com/BeneathHQ)

Beneath is a serverless DataOps platform that aims to combine data storage, processing, and visualization with data quality management and governance.

## Problem

The holy grail of data work is putting data science into production. It's glorious to behold a live dashboard that aggregates multiple data sources, to receive real-time alerts based on a machine learning model, or to offer customer-specific analytics in your frontend. Unfortunately, turning ad-hoc data science into a production data service also involves, in our humble opinion, an outright ridiculous engineering overhead.

For skilled data engineers, building a best-practices data management stack is literally a full time job. If we were starting a project from scratch today, we would set up Postgres, BigQuery, Kafka, Airflow, DBT and Metabase just to cover the basics. Then we'd start thinking about data quality management, data cataloging, data versioning, data lineage, permissions management, change data capture, stream processing, and so on, and so on.

For data scientists, many of these paradigms are completely unfamiliar. It's easy to stick with the tools you know. We've been there, and we wound up with a Postgres database and a hairball of Python code so unwieldy that not even we understood it anymore. We've interviewed a lot of small data science teams, and that's still the most common stack we see.

## Our solution

With Beneath, we're pursuing a unified approach that combines data storage, processing, and visualization with data quality management and governance in one platform. We're inspired by services like Netlify and Vercel that make it remarkable easy for developers to build and run web apps. In that same spirit, we're building a platform that data scientists and engineers can use to deploy, monitor, derive, visualize, integrate, and share analytics.

XXX

## Get started

The best way to try out Beneath is with a free beta account. When you sign up, you'll be guided to installing the SDK and setting up your first project. [Sign up here](https://beneath.dev/?noredirect=1).

We're working on bundling a self-hosted version that you can run locally. If you're interested in trying it out, [contact us](https://about.beneath.dev/contact) and let us know!

## Status

YYYY

<!-- Getting there is a journey and we're starting with data storage in the form of streams you can replay, subscribe to, query with SQL, monitor, and share. Streams are the primitive the rest of our roadmap builds upon. -->

## Features and roadmap

- **Data storage**
  - [x] Log streaming with at-least-once delivery
  - [x] Key-based table indexing for fast lookups
  - [x] Data warehouse replication for OLAP queries (SQL)
  - [x] Data versioning
  - [ ] Secondary table indexes
  - [ ] Schema evolution and migration
  - [ ] Log features: partitioning, compaction, multi-consumer subscriptions
  - [ ] Strongly consistent table operations for OLTP
  - [ ] Full-text search engine
  - [ ] Geo-replicated storage
- **Data processing**
  - [ ] (**In progress**) Scheduled/triggered SQL queries
  - [ ] Compute sandbox for batch and streaming pipelines
  - [ ] Git-based stream, query and pipeline deployments
  - [ ] Data app catalog (one-click parameterized deployments)
  - [ ] DAG view of streams and deriving pipelines for data lineage
- **Data visualization and exploration**
  - [ ] (**In progress**) Vega-based charts
  - [ ] (**In progress**) Dashboards composed from charts and tables
  - [ ] Alerting layer
  - [ ] Python notebooks (Jupyter)
- **Data quality and governance**
  - [x] Web console for creating and browsing resources
  - [x] Usage dashboards for streams, services, users and organizations
  - [x] Custom data usage quotas
  - [x] Granular permissions management, including public streams
  - [x] Service accounts with custom permissions and quotas
  - [x] API secrets (tokens) that can be issued/monitored/revoked
  - [ ] Field validation rules, checked on write
  - [ ] Data quality tests
  - [ ] Data search and discovery
  - [ ] Audit logs as meta-streams
- **Integrations**
  - [x] gRPC, REST and websockets APIs
  - [x] Command-line interface (CLI)
  - [x] Python client
  - [x] JS and React client
  - [ ] PostgreSQL wire-protocol compatibility
  - [ ] GraphQL API for data
  - [ ] Row restricted access tokens for identity-centered apps
  - [ ] Self-hosted Beneath on Kubernetes with federation

## How it works

ZZZ

## Community and Support

- Chat in [our Discord](https://discord.gg/f5yvx7YWau)
- Email us at [hello@beneath.dev](mailto:hello@beneath.dev)
- Book us for a casual [20-minute meeting](https://calendly.com/beneath-epg/beneath-office-hours)

## Documentation

- Homepage: [https://about.beneath.dev](https://about.beneath.dev)
- Documentation and tutorials: [https://about.beneath.dev/docs](https://about.beneath.dev/docs)
- Python client reference: [https://python.docs.beneath.dev](https://python.docs.beneath.dev)
- JavaScript client reference: [https://js.docs.beneath.dev](https://js.docs.beneath.dev)
- React client reference: [https://react.docs.beneath.dev](https://react.docs.beneath.dev)

## License

This repository contains the full source code for Beneath. It consists of an open-source core licensed under the MIT license, plus several source-available components found in the `ee` directory, most notably the Beneath UI. You can find more details in the `contributing/` directory.
