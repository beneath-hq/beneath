# `infra/`

This folder contains packages for connecting to the various *external* infrastructure components that Beneath relies upon, namely:

- `infra/db/` for connecting to Postgres
- `infra/engine/` for connecting to log, index and warehouse data services (including drivers for specific systems)
- `infra/mq/` for connecting to an at-least-once-delivery message queue (including drivers for specific systems)
- `infra/redis/` for connecting to Redis

