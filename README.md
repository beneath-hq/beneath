# README

## Git

Beneath aims to use a single git repository (monorepo) for all internal code. Hopefully that makes it easier to manage dependencies and get a full view of development.

Open source code (SDKs, libraries, examples) will have to be put in separate repositories. 

The project structure is currently:

- beneath-frontend contains JavaScript/TypeScript code for the frontend
- beneath-go contains all Go code, which spans the control plane, the data gateway and the core data pipelines

Checkout the git repo with: `git clone https://github.com/beneathcrypto/beneath-core.git`

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

- Go to `beneath-core/beneath-go`
- Copy `example.env` into `.env` and configure all variables (ask a colleague)
- Run with `go run cmd/control/main.go`
- If successful, you can access a GraphQL Playground at http://localhost:4000/playground

### Gateway backend

- Run with `go run cmd/gateway/main.go`

### Frontend (UI)

- Go to `beneath-core/beneath-frontend`
- Install dependencies with `yarn install` (if you don't have `yarn` installed: https://yarnpkg.com/en/docs/install)
- Run dev server with `yarn dev`


