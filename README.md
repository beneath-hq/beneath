<p align="center">
  <a href="https://about.beneath.dev/?utm_source=github&utm_medium=logo" target="_blank">
    <img src="https://raw.githubusercontent.com/beneath-hq/beneath/master/assets/logo/banner-icon-text-background.png" alt="Beneath" height="200">
  </a>
</p>

Beneath is a serverless real-time data platform. Our goal is to create one end-to-end platform for data workers that combines data storage, processing, and visualization with data quality management and governance.

<hr />

[![Go Report Card](https://goreportcard.com/badge/github.com/beneath-hq/beneath?style=flat-square)](https://goreportcard.com/report/github.com/beneath-hq/beneath)
[![GoDoc](https://godoc.org/github.com/beneath-hq/beneath?status.svg)](https://godoc.org/github.com/beneath-hq/beneath)
[![Twitter](https://img.shields.io/badge/Follow-BeneathHQ-blue.svg?style=flat&logo=twitter)](https://twitter.com/BeneathHQ)

_Beneath is a work in progress and your input makes a big difference! If you like it, star the project to show your support or [reach out](https://about.beneath.dev/contact/) and tell us what you think._

## 🧠 Philosophy

The holy grail of data work is putting data science into production. It's glorious to build live dashboards that aggregate multiple data sources, send real-time alerts based on a machine learning model, or offer customer-specific analytics in your frontend.

But building a modern data management stack is a full-time job, and a lot can go wrong. If you were starting a project from scratch today, you might set up Postgres, BigQuery, Kafka, Airflow, DBT and Metabase just to cover the basics. Later, you would need more tools to do data quality management, data cataloging, data versioning, data lineage, permissions management, change data capture, stream processing, and so on.

Beneath is a new way of building data apps. It takes an end-to-end approach that combines data storage, processing, and visualization with data quality management and governance in one serverless platform. The idea is to provide one opinionated layer of abstraction, i.e. one SDK and UI, which under the hood builds on modern data technologies.

Beneath is inspired by services like Netlify and Vercel that make it remarkable easy for developers to build and run web apps. In that same spirit, we want to give data scientists and engineers the fastest developer experience for building data products.

## 🚀 Status

We started with the data storage and governance layers. You can use the [Beneath Beta](https://beneath.dev/?noredirect=1) today to store, explore, query, stream, monitor and share data. It offers several interfaces, including a Python client, a CLI, websockets, and a web UI. The beta is stable for non-critical use cases. If you try out the beta and have any feedback to share, we'd love to [hear it](https://about.beneath.dev/contact/)!

Next up, we're tackling the data processing and data visualization layers, which will bring expanded opportunity for data governance and data quality management (see the roadmap at the end of this README for progress).

## 🎬 Tour

The snippet below presents a whirlwind tour of the Python API:

```python
# Create a new table
table = await client.create_table("examples/demo/foo", schema="""
  type Foo @schema {
    foo: String! @key
    bar: Timestamp
  }
""")

# Write batch or real-time data
await table.write(data)

# Load into a dataframe
df = await beneath.load_full(table)

# Replay and subscribe to changes
await beneath.consume(table, callback, subscription_path="...")

# Analyze with SQL
data = await beneath.query_warehouse(f"SELECT count(*) FROM `{table}`")

# Lookup by key, range or prefix
data = await table.query_index(filter={"foo": {"_prefix": "bar"}})
```

The image below shows a screenshot from the Beneath console. Check out the home page for a [demo video](https://about.beneath.dev/).

![Source code example](https://about.beneath.dev/media/readme/monitoring.png)

## 🐣 Get started

The best way to try Beneath is with a free beta account. [Sign up here](https://beneath.dev/?noredirect=1). When you have created an account, you can:

1. Install and authenticate the Beneath SDK
2. Browse public projects and integrate using Python, JavaScript, Websockets and more
3. Create a private or public project and start writing data

We're working on bundling a self-hosted version that you can run locally. If you're interested in self-hosting, [let us know](https://about.beneath.dev/contact)!

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

## 📦 Features and roadmap

- **Data storage**
  - [x] Log streaming for replay and subscribe
  - [x] Replication to key-value store for fast indexed lookups
  - [x] Replication to data warehouse for OLAP queries (SQL)
  - [x] Schema management and enforcement
  - [x] Data versioning
  - [ ] Schema evolution and migrations
  - [ ] Secondary indexes
  - [ ] Strongly consistent operations for OLTP
  - [ ] Geo-replicated storage
- **Data processing**
  - [ ] Scheduled/triggered SQL queries
  - [ ] Compute sandbox for batch and streaming pipelines
  - [ ] Git-integration for continuous deployments
  - [ ] DAG view of tables and pipelines for data lineage
  - [ ] Data app catalog (one-click parameterized deployments)
- **Data visualization and exploration**
  - [ ] Vega-based charts
  - [ ] Dashboards composed from charts and tables
  - [ ] Alerting layer
  - [ ] Python notebooks (Jupyter)
- **Data governance**
  - [x] Web console and CLI for creating and browsing resources
  - [x] Usage dashboards for tables, services, users and organizations
  - [x] Usage quota management
  - [x] Granular permissions management
  - [x] Service accounts with custom permissions and quotas
  - [x] API secrets (tokens) that can be issued/revoked
  - [ ] Data search and discovery
  - [ ] Audit logs as meta-tables
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

The `contributing/` directory in this repository contains a deeper technical walkthrough of the software architecture.

## 🛒 License

This repository contains the full source code for Beneath. Beneath's core is source available, licensed under the [Business Source License](https://github.com/beneath-hq/beneath/blob/master/licenses/BSL.txt), which converts to the Apache 2.0 license after four years. All the client libraries (in the `clients/` directory) and examples (in the `examples/` directory) are open-source, licensed under the MIT license.
