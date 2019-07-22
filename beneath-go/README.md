# beneath-go README

This directory contains all Go-related Beneath code. It spans multiple separate executable services, namely:

- The **control server**, which exposes a GraphQL API for managing users, projects, streams, etc.
- The **gateway server**, which exposes REST and gRPC APIs for reading and writing data to Beneath
- The **data processing pipeline**, which sits at the core of the platform and takes care of writing/forwarding data to the right places

## Project structure

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

## Control server

These are the main libraries used in the control server

- [`go-pg`](https://github.com/go-pg/pg): For interacting with the Postgres database. It's vaguely like an ORM. See the [wiki](https://github.com/go-pg/pg/wiki) for good examples on how to use it, especially the ["Model Definition"](https://github.com/go-pg/pg/wiki/Model-Definition) and ["Writing Queries"](https://github.com/go-pg/pg/wiki/Writing-Queries) pages. We're using [this helper library](https://github.com/go-pg/migrations) to run migrations (see below). 
- [`gqlgen`](https://gqlgen.com/): For defining the GraphQL server. It generates Go files based on GraphQL schema files (see below). Not the most common choice, but it seems many people think it's now the best GraphQL library for Go.
- [`goth`](https://github.com/markbates/goth): For setting up authentication with Google and Github
- [`chi`](https://github.com/go-chi/chi): For HTTP routing

### Adding GraphQL resolvers in the *control server*

1. Add/update the GraphQL schema files in `beneath-core/beneath-go/control/gql/schema`
2. Run `gqlgen-control.sh`
3. This is where it gets a little tricky -- it will generate stubs for all the resolvers in the file `beneath-core/beneath-go/control/resolver/generated.go`. You will want to pick out all the types/resolvers that have been updated/changed and add them to one of the other resolver files in the package (e.g. if it's a new user-related query, add it to `beneath-core/beneath-go/control/resolver/user`)
4. When you're done manually merging the new generated resolvers into the existing resolvers, delete the `generated.go` file. Good news: if you mess anything up, the type checker will complain!

### Running migrations

1. Add a migration file to `beneath-core/beneath-go/control/migrations`
2. Update the relevant model(s) in `beneath-core/beneath-go/control/model` to match the migration
3. The migration will automatically be applied when you start the control server
4. (There is also a migration tool, which you can run with `go run cmd/control_migrate/main.go XXX`, where XXX can be `up`, `down` and `reset` -- `reset` is especially useful during development)
