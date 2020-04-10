---
title: Setting up a development environment
description: 
menu:
  docs:
    parent: contributing
    weight: 300
weight: 300
---

In order to develop Beneath on your local machine, you need to setup local installations of its dependencies.

It's a good idea to run each dependency in a separate tab of a single terminal window.

## Installing and running dependencies

### Run Redis

- Install with `brew install redis`
- Run with `redis-server /usr/local/etc/redis.conf`

### Run Postgres

- We suggest you install https://postgresapp.com/ and run it in the background. It's useful to keep an open `psql` terminal tab to query the database directly.
- Optionally, on Macs, we like using Postico, which you can find here: https://eggerapps.at/postico/

### Install Google Cloud SDK

Follow this tutorial https://cloud.google.com/sdk/docs/downloads-interactive, but first read this: It creates a folder in the directory from which you run the install commands, so make sure you're in a folder where you won't delete it by accident (probably home or documents). 

Check that everything installed correctly and that you're in the `beneath` Google Cloud project by running `gcloud projects list` (in a new tab). You might also want to check out your `~/.bash_profile` to make sure it configured your `PATH` correctly.

### Run Cloud Pubsub emulator

- Follow install instructions here: https://cloud.google.com/pubsub/docs/emulator
- Run with `gcloud beta emulators pubsub start`

### Run Cloud Bigtable emulator

- Follow install instructions here: https://cloud.google.com/bigtable/docs/emulator
- Run with `gcloud beta emulators bigtable start`

### Install VS Code

- If you can stand it, we recommend you use VS Code: https://code.visualstudio.com/
- Also install the CLI shortcut: Hit Shift-Command-P and search for "Shell Command: Install 'code' command in PATH
- Install the Go extension from Microsoft and all of its Go Tool dependencies: https://github.com/microsoft/vscode-go#how-to-use-this-extension

## Running Beneath locally

### Configure environment variables

- Copy `configs/example.env` into `configs/.development.env` and `configs/.test.env` and configure all variables (ask a core team member)

### Control backend

- For development, run with `ENV=dev go run cmd/control/main.go`
- If successful, you can access a GraphQL Playground at http://localhost:4000/playground

### Gateway backend

- Run with `ENV=dev go run cmd/gateway/main.go`
- If successful, you can ping the gateway at http://localhost:5000/

### Pipeline

- Run with `ENV=dev go run cmd/pipeline/main.go`

### Task queue

- Run with `ENV=dev go run cmd/taskqueue/main.go`

### Frontend (UI)

- Go to `web/`
- Install dependencies with `yarn install` (if you don't have `yarn` installed, [go here](https://yarnpkg.com/en/docs/install))
- Run dev server with `yarn dev`
