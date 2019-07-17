# beneath-go README

This directory contains all Go-related Beneath code. It spans multiple separate executable services, namely:

- The **control server**, which exposes a GraphQL API for managing users, projects, streams, etc.
- The **gateway server**, which exposes REST and gRPC APIs for reading and writing data to Beneath
- The **data processing pipeline**, which sits at the core of the platform and takes care of writing/forwarding data to the right places

### Project structure

Here's a rough guide to the project structure:

- `cmd`: contains a package with a `main.go` file for each executable program
- `control`: code related to the control server, including Postgres models
- `core`: utility libraries that are not directly related to any specific service (e.g., the schema parser)
- `engine`: code for interfacing with the underlying data infrastructure, e.g., Pubsub, BigTable, BigQuery, etc.
- `gateway`: code related to the gateway server (excluding data infrastructure code -- that's in `engine`)
- `proto`: [protocol buffers](https://developers.google.com/protocol-buffers/) files, which for now we keep separate from their logical place in the architecture

### Helper scripts

For now, we put helper scripts right in the `beneath-go` folder. Guide to helper scripts:

- `gqlgen-control.sh`: (re)generates the control server's GraphQL resolvers (see its README for more)
- `proto-build.sh`: (re)generates the protocol buffers classes
