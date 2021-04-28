---
title: Project structure
description:
menu:
  docs:
    parent: contributing
    weight: 200
weight: 200
---

## Monorepo

Beneath aims to use a single git repository (monorepo) for all code and documentation. Hopefully that makes it easier to manage dependencies and get a full view of development.

Checkout the git repo with: `git clone https://github.com/beneath-hq/beneath.git`

Client libraries (i.e. language-specific libraries that read from and write to Beneath) are found in the `clients/` directory. As an example, see the [Python client library](https://github.com/beneath-hq/beneath/tree/master/clients/python).

## Core components

The entire backend is implemented in Go, the frontend is implemented in TypeScript (React), and the client libraries are (naturally) implemented in a variety of different languages.

Conceptually, the backend is divided into two planes:

- The **control plane** manages users, projects, streams, etc., and is backed by Postgres and Redis.
- The **data plane** reads and writes stream records, and is backed by pluggable data systems for log storage, indexed lookups and warehouse/OLAP queries (known as the "engine", see `infra/engine/`).

The backend is compiled as one executable (`cmd/beneath/`) that can run one or all of the following services:

- `control-server`, which exposes a GraphQL API for the control plane
- `control-worker`, which runs background tasks related to the control plane
- `data-server`, which exposes REST, Websockets and GRPC APIs for the data plane (i.e. reading and writing streams)
- `data-worker`, which asynchronously processes records written to `data-server` and writes them to the underlying data systems

Here's a rough guide to the project structure (each folder has a `README.md` with further details):

- `assets`: Assets such as logos, images, etc.
- `build`: Dockerfiles for the deployed executables in `cmd`
- `bus`: Control-plane message bus for in-process and async event publish-subscribe; publishers and subscribers are found in `services/`, and events are defined in `models/`
- `clients`: Language-specific libraries used to interface with Beneath, including reading and writing data
- `cmd`: Contains the `main.go` for the backend executable
- `config`: Contains a YAML config template, and may contain local development config (e.g. `.development.yaml`, which is ignored by Git)
- `contributing`: Documentation that describes how the codebase is structured and how to contribute to it
- `ee`: Enterprise-edition functionality, such as billing
- `infra`: Packages for connecting to external systems / infrastructure, such as Postgres, Redis, MQ, and notably the "engine", which has drivers for the data systems used in the data-plane
- `migrations`: Postgres migrations for the control-plane models in `models/`
- `models`: Control-plane models and bus events
- `pkg`: Stand-alone utility libraries that are not directly related to any specific service (e.g., the stream schema parser)
- `scripts`: Helper scripts for code generation, etc.
- `server`: Implementations of the control and data servers
- `services`: Contains packages ("services") that encapsulate functionality for a specific domain of the app, typically involving persistant operations on `infra/`
- `test`: End-to-end and integration tests
- `web`: Contains the frontend (data console) code

## Technologies

- We use Go for all backend-related code
- We use TypeScript and React (Next.js) for the frontend (found in `web/`)
- We use GraphQL for the control plane and [gRPC](https://grpc.io/)/[Protocol buffers](https://developers.google.com/protocol-buffers/) for the data plane
- We use Postgres and Redis as the underlying data stores for the control plane
- The data plane supports several backends. Our hosted offering runs on Google BigTable, Google PubSub and Google BigQuery.
