# `cmd/beneath/`

This folder contains the main executable for the Beneath backend. It can start different backend services and manage migrations. Run `BENEATH_ENV=dev go run cmd/beneath/main.go --help` for a quick overview.

**To start the backend in development, run: `BENEATH_ENV=dev go run cmd/beneath/main.go start all`**

This folder also contains config management and dependency initialization/injection code. The implementation may seem unexpectedly complicated, but it's necessary to for two reasons:

1. The CLI and dependencies are extended/overwritten by the Enterprise executable in `ee/cmd/beneath/`
2. We support running the different "startables" standalone or all in one process

We use [Cobra](https://github.com/spf13/cobra) for the CLI, [Viper](https://github.com/spf13/viper) for configuration and [Dig](https://github.com/uber-go/dig) for dependency injection/initialization. 

## Overview

- `cmd/beneath/cli/` contains generic functionality for creating the CLI, dependency initialization/injection and application lifecycle management.
- `cmd/beneath/dependencies/` registers dependency constructors and config keys.
- `cmd/beneath/main.go` defines CLI subcommands and startable services.

## Concepts 

The following concepts/lingo are useful when editing the code.

- `Startable`: Services that can be started with the CLI
  - Started with `beneath start NAME` (or automatically when running `beneath start all`)
  - Take dependencies as arguments, which are automatically (and lazily) initialized
  - Must take `*cli.Lifecycle` as an argument (and use it to actually run services)

- `Dependency`: Constructors that build dependencies, optionally taking other dependencies as arguments
  - We use reflection-based dependency injection. We use [Dig](https://github.com/uber-go/dig) under the hood, which basically form a DAG based on the input and output types of the registered dependencies

- `Lifecycle`: Tracks services to run, and runs them concurrently in one process. It provides running services with a `context` that is cancelled on SIGINT/SIGTERM or when another service returns an error. It is inspired by [fx.Lifecycle](https://pkg.go.dev/go.uber.org/fx), but we prefer a `Run(ctx context.Context) error` signature to a start/stop hooks-based approach.

- `ConfigKey`: Specifies a configuration key with optional default value and description (see `config/` for more info)
