# DEVELOPING

## Git

Beneath aims to use a single git repository (monorepo) for all internal code. Hopefully that makes it easier to manage dependencies and get a full view of development.

Open source code (SDKs, libraries, examples) will have to be put in separate repositories. 

The project structure is currently:

- web contains JavaScript/TypeScript code for the frontend
- beneath-go contains all Go code, which spans the control plane, the data gateway and the core data pipelines

Checkout the git repo with: `git clone https://gitlab.com/_beneath/beneath-core.git`

## Setting up a development environment

In order to develop Beneath on your local machine, you need to setup local installations of its dependencies.

It's a good idea to run each dependency in a separate tab of a single terminal window.

### Run Redis

- Install with `brew install redis`
- Run with `redis-server /usr/local/etc/redis.conf`

### Run Postgres

We suggest you install https://postgresapp.com/ and run it in the background. It's useful to keep an open `psql` terminal tab to query the database directly.

### Install Google Cloud SDK

Follow this tutorial https://cloud.google.com/sdk/docs/downloads-interactive, but first read this: It creates a folder in the directory from which you run the install commands, so make sure you're in a folder where you won't delete it by accident (probably home or documents). 

Check that everything installed correctly and that you're in the `beneathcrypto` Google Cloud project by running `gcloud projects list` (in a new tab). You might also want to check out your `~/.bash_profile` to make sure it configured your `PATH` correctly.

### Run Cloud Pubsub emulator

- Follow install instructions here: https://cloud.google.com/pubsub/docs/emulator
- Run with `gcloud beta emulators pubsub start`

### Run Cloud Bigtable emulator

- Follow install instructions here: https://cloud.google.com/bigtable/docs/emulator
- Run with `gcloud beta emulators bigtable start`

### Install VS Code

- If you can stand it, I recommend you use VS Code: https://code.visualstudio.com/
- Also install the CLI shortcut: Hit Shift-Command-P and search for "Shell Command: Install 'code' command in PATH

## Running Beneath locally

### Control backend

- Go to `beneath-core`
- Copy `example.env` into `.env` and configure all variables (ask a colleague)
- Run with `go run cmd/control/main.go`
- If successful, you can access a GraphQL Playground at http://localhost:4000/playground

### Gateway backend

- Go to `beneath-core`
- Run with `go run cmd/gateway/main.go`
- If successful, you can ping the gateway at http://localhost:5000/

### Frontend (UI)

- Go to `beneath-core/web`
- Install dependencies with `yarn install` (if you don't have `yarn` installed: https://yarnpkg.com/en/docs/install)
- Run dev server with `yarn dev`

## Project structure

This directory contains all Go-related Beneath code. It spans multiple separate executable services, namely:

- The **control server**, which exposes a GraphQL API for managing users, projects, streams, etc.
- The **gateway server**, which exposes REST and gRPC APIs for reading and writing data to Beneath
- The **data processing pipeline**, which sits at the core of the platform and takes care of writing/forwarding data to the right places

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

- [`go-pg`](https://github.com/go-pg/pg): For interacting with the Postgres database. It's vaguely like an ORM. See the [wiki](https://github.com/go-pg/pg/v9/wiki) for good examples on how to use it, especially the ["Model Definition"](https://github.com/go-pg/pg/v9/wiki/Model-Definition) and ["Writing Queries"](https://github.com/go-pg/pg/v9/wiki/Writing-Queries) pages. We're using [this helper library](https://github.com/go-pg/migrations/v7) to run migrations (see below). 
- [`gqlgen`](https://gqlgen.com/): For defining the GraphQL server. It generates Go files based on GraphQL schema files (see below). Not the most common choice, but it seems many people think it's now the best GraphQL library for Go.
- [`goth`](https://github.com/markbates/goth): For setting up authentication with Google and Github
- [`chi`](https://github.com/go-chi/chi): For HTTP routing

### Adding GraphQL resolvers in the *control server*

1. Add/update the GraphQL schema files in `beneath-core/control/gql/schema`
2. Run `gqlgen-control.sh`
3. This is where it gets a little tricky -- it will generate stubs for all the resolvers in the file `beneath-core/control/resolver/generated.go`. You will want to pick out all the types/resolvers that have been updated/changed and add them to one of the other resolver files in the package (e.g. if it's a new user-related query, add it to `beneath-core/control/resolver/user`)
4. When you're done manually merging the new generated resolvers into the existing resolvers, delete the `generated.go` file. Good news: if you mess anything up, the type checker will complain!

### Running migrations

1. Add a migration file to `beneath-core/control/migrations`
2. Update the relevant model(s) in `beneath-core/control/entity` to match the migration
3. The migration will automatically be applied when you start the control server
4. (There is also a migration tool, which you can run with `go run cmd/control_migrate/main.go XXX`, where XXX can be `up`, `down` and `reset` -- `reset` is especially useful during development)
