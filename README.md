<p align="center">
  <a href="https://about.beneath.dev/?utm_source=github&utm_medium=logo" target="_blank">
    <img src="https://raw.githubusercontent.com/beneath-hq/beneath/master/assets/logo/banner-icon-text-background.png" alt="Beneath" height="200">
  </a>
</p>

Beneath is a serverless DataOps platform. It aims to combine data storage, processing, and visualization with data quality management and governance in one end-to-end platform.

<hr />

[![Go Report Card](https://goreportcard.com/badge/gitlab.com/beneath-hq/beneath?style=flat-square)](https://goreportcard.com/report/gitlab.com/beneath-hq/beneath)
[![GoDoc](https://godoc.org/gitlab.com/beneath-hq/beneath?status.svg)](https://godoc.org/gitlab.com/beneath-hq/beneath)
[![Twitter](https://img.shields.io/badge/Follow-BeneathHQ-blue.svg?style=flat&logo=twitter)](https://twitter.com/BeneathHQ)

_Beneath is a work in progress and your input makes a big difference! Star the project to show your support or [reach out](https://about.beneath.dev/contact) and tell us what you think._

## 🧠 Philosophy

The holy grail of data work is putting data science into production. It's glorious to behold a live dashboard that aggregates multiple data sources, send real-time alerts based on a machine learning model, or offer customer-specific analytics in your frontend.

But building a modern data management stack is a full-time job, and a lot can go wrong. If you were starting a project from scratch today, you might set up Postgres, BigQuery, Kafka, Airflow, DBT and Metabase just to cover the basics. Later, you would need more tools to do data quality management, data cataloging, data versioning, data lineage, permissions management, change data capture, stream processing, and so on, and so on.

Beneath is a new way of building data apps. It takes an end-to-end approach that combines data storage, processing, and visualization with data quality management and governance in one platform. The idea is to provide one opinionated layer of abstraction, which under the hood builds on several modern data technologies and frameworks.

Beneath is inspired by services like Netlify and Vercel that make it remarkable easy for developers to build and run web apps. In that same spirit, we want to give data scientists and engineers a better developer experience to deploy, monitor, enrich, visualize, integrate, and share analytics.

## 🐣 Get started

The best way to try Beneath is with a free beta account. When you sign up, you'll be guided to installing the Python SDK and setting up your first project. [Sign up here](https://beneath.dev/?noredirect=1).

We're working on bundling a self-hosted version that you can run locally. If you're interested in self-hosting, [let us know](https://about.beneath.dev/contact)!

## 📦 Features and roadmap

**Status:** The data storage and governance layers are currently [in beta](https://beneath.dev/?noredirect=1), and stable for non-critical use cases. We're ramping up to add data processing and data visualization capabilities, which in turn will unlock the opportunity for more advanced data quality management.

- **Data storage**
  - [x] Log streaming for replay and subscribe
  - [x] Replication to key-value store for fast indexed lookups
  - [x] Replication to data warehouse for OLAP queries (SQL)
  - [x] Schema management and enforcement
  - [x] Data versioning
  - [ ] Secondary indexes
  - [ ] Schema evolution and migrations
  - [ ] Strongly consistent operations for OLTP
  - [ ] Full-text search engine
  - [ ] Geo-replicated storage
- **Data processing**
  - [ ] (**In progress**) Scheduled/triggered SQL queries
  - [ ] Compute sandbox for batch and streaming pipelines
  - [ ] Git-based stream, query and pipeline deployments
  - [ ] Data app catalog (one-click parameterized deployments)
  - [ ] DAG view of streams and deriving pipelines for data lineage
- **Data visualization and exploration**
  - [ ] Vega-based charts
  - [ ] Dashboards composed from charts and tables
  - [ ] Alerting layer
  - [ ] Python notebooks (Jupyter)
- **Data governance**
  - [x] Web console and CLI for creating and browsing resources
  - [x] Usage dashboards for streams, services, users and organizations
  - [x] Usage quota management
  - [x] Granular permissions management
  - [x] Service accounts with custom permissions and quotas
  - [x] API secrets (tokens) that can be issued/revoked
  - [ ] Data search and discovery
  - [ ] Audit logs as meta-streams
- **Data quality management**
  - [ ] Field validation rules, checked on write
  - [ ] Alert triggers
  - [ ] Data distribution tests
  - [ ] Machine learning model re-training and monitoring
- **Integrations**
  - [x] gRPC, REST and websockets APIs
  - [x] Command-line interface (CLI)
  - [x] Python client
  - [x] JS and React client
  - [ ] PostgreSQL wire-protocol compatibility
  - [ ] GraphQL API for data
  - [ ] Row restricted access tokens for identity-centered apps
  - [ ] Self-hosted Beneath on Kubernetes with federation

## 🍿 How it works

Check out the [Concepts](https://about.beneath.dev/docs/concepts/) section of the docs for an overview of how Beneath works.

The `contributing/` directory in this repository contains a deeper technical walkthrough of the technical architecture.

## 👋 Community and Support

- Chat in [our Discord](https://discord.gg/f5yvx7YWau)
- Email us at [hello@beneath.dev](mailto:hello@beneath.dev)
- Book a casual [20-minute meeting](https://calendly.com/beneath-epg/beneath-office-hours)

## 🎓 Documentation

- Homepage: [https://about.beneath.dev](https://about.beneath.dev)
- Documentation and tutorials: [https://about.beneath.dev/docs](https://about.beneath.dev/docs)
- Python client reference: [https://python.docs.beneath.dev](https://python.docs.beneath.dev)
- JavaScript client reference: [https://js.docs.beneath.dev](https://js.docs.beneath.dev)
- React client reference: [https://react.docs.beneath.dev](https://react.docs.beneath.dev)

## 🛒 License

This repository contains the full source code for Beneath. It consists of an open-source core licensed under the MIT license, plus several source-available components found in the `ee` directory, most notably the Beneath UI. The source-available components are free to use unless you intend to offer Beneath as a service. You can find more details in the `contributing/` directory.
