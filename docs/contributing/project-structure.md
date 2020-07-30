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

Checkout the git repo with: `git clone https://gitlab.com/beneath-hq/beneath.git`

Client libraries (i.e. language-specific libraries that read from and write to Beneath) are found in the `clients/` directory. As an example, see the [Python client library](https://gitlab.com/beneath-hq/beneath/-/tree/master/clients/python).

## Core components

The entire system is implemented in Go with the exception of the frontend, which is implemented in TypeScript (and React), and the client libraries, which are implemented for a variety of different languages.

The core system spans multiple separate executable services, namely:

- The **control server** (found in `control/`), which manages users, projects, streams, etc. It exposes a GraphQL API and is backed by Postgres and Redis.
- The **task queue** (found in `control/taskqueue/`), which executes background jobs issued by the control server.
- The **gateway server** (found in `gateway/`), which reads and writes data to Beneath. It exposes REST and gRPC APIs and interfaces with the underlying infrastructure using its **pipeline** and the **engine** (found in `engine/`).
- The **pipeline** (found in `gateway/pipeline/`), which in the background receives data from the gateway and writes it to the engine.

Here's a rough guide to the project structure:

- `assets`: Assets such as logos, images, etc.
- `build`: Dockerfiles for the deployed executables in `cmd`
- `clients`: Language-specific libraries used to interface with Beneath, including reading and writing data
- `cmd`: contains a package with a `main.go` file for each executable program
- `configs`: should contain `.development.env` and `.test.env` files that configure the platform in local development
- `control`: code related to the control server, including Postgres models, GraphQL resolvers, and the task queue
- `deployments`: Kubernetes and Helm manifests for deploying to production
- `docs/contributing`: documentation that describes how the codebase is structured and how to contribute to it
- `engine`: drivers for interfacing with the underlying data infrastructure, e.g., Pubsub, BigTable, BigQuery, etc.
- `gateway`: code related to the data gateway server (excluding data infrastructure code -- that's in `engine`)
- `internal`: utility libraries that are not directly related to any specific service (e.g., HTTP middleware)
- `pkg`: stand-alone utility libraries that are not directly related to any specific service (e.g., the stream schema parser)
- `scripts`: contains helper scripts for code generation, etc.
- `test`: contains end-to-end and integration tests
- `web`: contains the frontend (data console) code

## Technologies

- We use Go for all backend-related code
- We use TypeScript and React (Next.js) for the frontend (found in `web/`)
- We use GraphQL for the control plane and [gRPC](https://grpc.io/)/[Protocol buffers](https://developers.google.com/protocol-buffers/) for the data plane
- We use Postgres and Redis as the underlying data stores for the control plane
- The data plane supports several backends. Our hosted offering runs on Google BigTable, Google PubSub and Google BigQuery.
