# `models/`

This folder contains structs for database models and bus events. It also contains simple implementation code, e.g. for validation or for adhering to interfaces, but it should not contain complex functionality (certainly not code that calls services or infrastructure).

`bus.Bus` marshals async events with `msgpack`, so every struct used in events should have sensible `msgpack` struct tags (eg. use `msgpack:"-"` for sub-objects like database relationships).

## Adding and running migrations

When you add or change a model, add related migrations to `migrations/`. Currently, migrations run automatically when you start the control server.

## Database library

We use [`go-pg`](https://github.com/go-pg/pg) for mapping structs to the Postgres database (in `infra/db/`). It's vaguely like an ORM. See the [wiki](https://github.com/go-pg/pg/v9/wiki) for good examples on how to use it, especially the ["Model Definition"](https://github.com/go-pg/pg/v9/wiki/Model-Definition) and ["Writing Queries"](https://github.com/go-pg/pg/v9/wiki/Writing-Queries) pages. We're using [this helper library](https://github.com/go-pg/migrations/v7) to run migrations (see below). 

